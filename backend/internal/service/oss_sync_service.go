package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gorm.io/gorm"
)

const ossSourceType = "oss"

// OSSSyncResult 用来返回一次同步任务的执行结果。
type OSSSyncResult struct {
	Trigger       string    `json:"trigger"`
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	ScannedCount  int       `json:"scanned_count"`
	CreatedCount  int       `json:"created_count"`
	UpdatedCount  int       `json:"updated_count"`
	RestoredCount int       `json:"restored_count"`
	DeletedCount  int       `json:"deleted_count"`
	SkippedCount  int       `json:"skipped_count"`
}

// OSSSyncService 负责把阿里云 OSS 里的 Markdown 同步到博客系统。
type OSSSyncService struct {
	config       config.AppConfig
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository

	client *oss.Client
	bucket *oss.Bucket

	mutex      sync.RWMutex
	lastResult OSSSyncResult
	isRunning  bool
}

// NewOSSSyncService 创建 OSS 同步服务。
func NewOSSSyncService(
	appConfig config.AppConfig,
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) (*OSSSyncService, error) {
	service := &OSSSyncService{
		config:       appConfig,
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		lastResult: OSSSyncResult{
			Success: false,
			Message: "还没有执行过同步任务。",
		},
	}

	if !service.IsEnabled() {
		service.lastResult.Message = "OSS 同步未启用。"
		return service, nil
	}

	client, err := oss.New(
		"https://"+strings.TrimSpace(appConfig.OSS.Endpoint),
		strings.TrimSpace(appConfig.OSS.AccessKeyID),
		strings.TrimSpace(appConfig.OSS.AccessKeySecret),
		oss.Region(strings.TrimSpace(appConfig.OSS.Region)),
	)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(strings.TrimSpace(appConfig.OSS.Bucket))
	if err != nil {
		return nil, err
	}

	service.client = client
	service.bucket = bucket
	service.lastResult.Message = "OSS 同步已启用，等待首次执行。"

	return service, nil
}

// IsEnabled 用来判断当前是否已经具备执行同步的必要配置。
func (service *OSSSyncService) IsEnabled() bool {
	return service.config.Sync.Enabled &&
		strings.TrimSpace(service.config.OSS.Endpoint) != "" &&
		strings.TrimSpace(service.config.OSS.Region) != "" &&
		strings.TrimSpace(service.config.OSS.AccessKeyID) != "" &&
		strings.TrimSpace(service.config.OSS.AccessKeySecret) != "" &&
		strings.TrimSpace(service.config.OSS.Bucket) != ""
}

// Status 返回当前同步状态，供后台页面展示。
func (service *OSSSyncService) Status() map[string]any {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	return map[string]any{
		"enabled":          service.IsEnabled(),
		"is_running":       service.isRunning,
		"bucket":           service.config.OSS.Bucket,
		"endpoint":         service.config.OSS.Endpoint,
		"region":           service.config.OSS.Region,
		"prefix":           normalizeOSSPrefix(service.config.OSS.Prefix),
		"interval_minutes": service.effectiveIntervalMinutes(),
		"last_result":      service.lastResult,
	}
}

// StartAutoSync 启动后台定时同步任务。
// 为了让系统一启动就有内容，这里会先异步跑一次 startup 同步。
func (service *OSSSyncService) StartAutoSync() {
	if !service.IsEnabled() {
		log.Println("OSS 同步未启用，跳过定时任务启动。")
		return
	}

	go func() {
		if _, err := service.RunOnce(context.Background(), "startup"); err != nil {
			log.Printf("首次 OSS 同步失败: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(service.effectiveIntervalMinutes()) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := service.RunOnce(context.Background(), "schedule"); err != nil {
				log.Printf("定时 OSS 同步失败: %v", err)
			}
		}
	}()
}

// RunOnce 手动执行一次同步。
func (service *OSSSyncService) RunOnce(_ context.Context, trigger string) (OSSSyncResult, error) {
	finalize := func(result OSSSyncResult, err error) (OSSSyncResult, error) {
		result.FinishedAt = time.Now()
		service.setLastResult(result, false)
		return result, err
	}

	if !service.IsEnabled() {
		result := OSSSyncResult{
			Trigger:   trigger,
			StartedAt: time.Now(),
			Success:   false,
			Message:   "OSS 同步未启用，请先检查配置。",
		}
		return finalize(result, errors.New(result.Message))
	}

	service.mutex.Lock()
	if service.isRunning {
		lastResult := service.lastResult
		service.mutex.Unlock()
		return finalize(lastResult, errors.New("同步任务正在执行中，请稍后再试"))
	}

	service.isRunning = true
	service.lastResult = OSSSyncResult{
		Trigger:   trigger,
		StartedAt: time.Now(),
		Success:   false,
		Message:   "同步任务正在执行中。",
	}
	service.mutex.Unlock()

	startedAt := time.Now()
	result := OSSSyncResult{
		Trigger:   trigger,
		StartedAt: startedAt,
		Success:   false,
		Message:   "同步开始执行。",
	}

	objectKeys, err := service.listMarkdownObjectKeys()
	if err != nil {
		result.Message = "读取 OSS 文件列表失败。"
		return finalize(result, err)
	}

	seenSourceKeys := make(map[string]struct{}, len(objectKeys))

	for _, objectKey := range objectKeys {
		result.ScannedCount++

		importedData, skipReason, err := service.loadImportedArticleData(objectKey)
		if err != nil {
			result.Message = "读取 Markdown 文件失败。"
			return finalize(result, err)
		}

		if skipReason != "" {
			result.SkippedCount++
			log.Printf("跳过 OSS 文件 %s: %s", objectKey, skipReason)
			continue
		}

		seenSourceKeys[importedData.SourceKey] = struct{}{}

		action, err := service.upsertImportedArticle(importedData)
		if err != nil {
			result.Message = "写入文章失败。"
			return finalize(result, err)
		}

		switch action {
		case "created":
			result.CreatedCount++
		case "updated":
			result.UpdatedCount++
		case "restored":
			result.RestoredCount++
		case "unchanged":
		default:
			result.SkippedCount++
		}
	}

	deletedCount, err := service.softDeleteMissingArticles(seenSourceKeys)
	if err != nil {
		result.Message = "处理已删除 OSS 文件失败。"
		return finalize(result, err)
	}

	result.DeletedCount = deletedCount
	result.Success = true
	result.Message = fmt.Sprintf(
		"同步完成，共扫描 %d 篇 Markdown，新建 %d 篇，更新 %d 篇，恢复 %d 篇，软删除 %d 篇，跳过 %d 篇。",
		result.ScannedCount,
		result.CreatedCount,
		result.UpdatedCount,
		result.RestoredCount,
		result.DeletedCount,
		result.SkippedCount,
	)

	return finalize(result, nil)
}

// listMarkdownObjectKeys 负责列出 blog 前缀下所有 Markdown 文件。
func (service *OSSSyncService) listMarkdownObjectKeys() ([]string, error) {
	prefix := normalizeOSSPrefix(service.config.OSS.Prefix)
	marker := ""
	keys := make([]string, 0)

	for {
		result, err := service.bucket.ListObjects(
			oss.Prefix(prefix),
			oss.Marker(marker),
			oss.MaxKeys(1000),
		)
		if err != nil {
			return nil, err
		}

		for _, object := range result.Objects {
			lowerKey := strings.ToLower(object.Key)
			if !strings.HasSuffix(lowerKey, ".md") {
				continue
			}
			keys = append(keys, object.Key)
		}

		if !result.IsTruncated {
			break
		}

		marker = result.NextMarker
	}

	return keys, nil
}

// loadImportedArticleData 负责下载并解析一篇 Markdown。
func (service *OSSSyncService) loadImportedArticleData(objectKey string) (importedArticleData, string, error) {
	relativeKey := strings.TrimPrefix(strings.Trim(strings.TrimSpace(objectKey), "/"), normalizeOSSPrefix(service.config.OSS.Prefix))
	relativeKey = strings.Trim(relativeKey, "/")
	parts := strings.Split(relativeKey, "/")

	// 这里要求目录结构必须是 blog/分类/文章.md。
	if len(parts) != 2 {
		return importedArticleData{}, "目录层级不是“blog/分类/文章.md”", nil
	}

	contentReader, err := service.bucket.GetObject(objectKey)
	if err != nil {
		return importedArticleData{}, "", err
	}
	defer contentReader.Close()

	rawBytes, err := io.ReadAll(contentReader)
	if err != nil {
		return importedArticleData{}, "", err
	}

	return parseImportedMarkdown(service.config.OSS, objectKey, string(rawBytes)), "", nil
}

// upsertImportedArticle 根据 source_key 来决定是新建文章还是更新文章。
func (service *OSSSyncService) upsertImportedArticle(importedData importedArticleData) (string, error) {
	category, err := service.categoryRepo.FirstOrCreate(importedData.Category)
	if err != nil {
		return "", err
	}

	tags, err := service.buildTags(importedData.TagNames)
	if err != nil {
		return "", err
	}

	now := time.Now()
	existingArticle, err := service.articleRepo.FindBySourceKey(ossSourceType, importedData.SourceKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newArticle := model.Article{
			Title:        importedData.Title,
			Content:      importedData.Content,
			Summary:      importedData.Summary,
			CoverURL:     importedData.CoverURL,
			Status:       1,
			CategoryID:   &category.ID,
			Tags:         tags,
			SourceType:   ossSourceType,
			SourceKey:    importedData.SourceKey,
			SourcePath:   importedData.SourcePath,
			SourceHash:   importedData.SourceHash,
			LastSyncedAt: &now,
		}

		if err := service.articleRepo.Create(&newArticle); err != nil {
			return "", err
		}

		return "created", nil
	}

	oldHash := existingArticle.SourceHash
	existingArticle.Title = importedData.Title
	existingArticle.Content = importedData.Content
	existingArticle.Summary = importedData.Summary
	existingArticle.CoverURL = importedData.CoverURL
	existingArticle.Status = 1
	existingArticle.CategoryID = &category.ID
	existingArticle.Tags = tags
	existingArticle.SourcePath = importedData.SourcePath
	existingArticle.SourceHash = importedData.SourceHash
	existingArticle.LastSyncedAt = &now

	wasDeleted := existingArticle.DeletedAt.Valid
	contentChanged := oldHash != importedData.SourceHash || wasDeleted

	if wasDeleted {
		if err := service.articleRepo.Restore(existingArticle.ID); err != nil {
			return "", err
		}
	}

	if err := service.articleRepo.Update(&existingArticle); err != nil {
		return "", err
	}

	if wasDeleted {
		return "restored", nil
	}

	if contentChanged {
		return "updated", nil
	}

	return "unchanged", nil
}

// buildTags 确保 frontmatter 里出现的标签都存在。
func (service *OSSSyncService) buildTags(tagNames []string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0, len(tagNames))

	for _, tagName := range tagNames {
		trimmedName := strings.TrimSpace(tagName)
		if trimmedName == "" {
			continue
		}

		tag, err := service.tagRepo.FirstOrCreate(trimmedName)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// softDeleteMissingArticles 负责把“OSS 上已经不存在”的文章做软删除。
func (service *OSSSyncService) softDeleteMissingArticles(seenSourceKeys map[string]struct{}) (int, error) {
	existingArticles, err := service.articleRepo.ListBySourceType(ossSourceType)
	if err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, article := range existingArticles {
		if _, exists := seenSourceKeys[article.SourceKey]; exists {
			continue
		}

		if article.DeletedAt.Valid {
			continue
		}

		if err := service.articleRepo.Delete(article.ID); err != nil {
			return deletedCount, err
		}

		deletedCount++
	}

	return deletedCount, nil
}

// effectiveIntervalMinutes 返回一个安全的同步间隔。
func (service *OSSSyncService) effectiveIntervalMinutes() int {
	if service.config.Sync.IntervalMinutes <= 0 {
		return 60
	}

	return service.config.Sync.IntervalMinutes
}

// setLastResult 统一更新运行状态和上次结果。
func (service *OSSSyncService) setLastResult(result OSSSyncResult, isRunning bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.lastResult = result
	service.isRunning = isRunning
}

// normalizeOSSPrefix 确保前缀一定带有结尾斜杠。
func normalizeOSSPrefix(prefix string) string {
	normalized := strings.Trim(strings.TrimSpace(prefix), "/")
	if normalized == "" {
		return ""
	}

	return path.Clean(normalized) + "/"
}
