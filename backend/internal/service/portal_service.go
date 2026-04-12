package service

import (
	"strings"
	"time"

	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"
)

// PortalService 负责前台读者端需要的数据。
type PortalService struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
	commentRepo  repository.CommentRepository
}

// NewPortalService 创建前台内容服务。
func NewPortalService(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	commentRepo repository.CommentRepository,
) *PortalService {
	return &PortalService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		commentRepo:  commentRepo,
	}
}

// ListCategories 返回分类列表。
func (service *PortalService) ListCategories() ([]model.Category, error) {
	return service.categoryRepo.List()
}

// ListTags 返回标签列表。
func (service *PortalService) ListTags() ([]model.Tag, error) {
	return service.tagRepo.List()
}

// ListArchives 返回归档数据。
func (service *PortalService) ListArchives() ([]map[string]any, error) {
	articles, err := service.articleRepo.ListPublished("")
	if err != nil {
		return nil, err
	}

	type archiveBucket struct {
		Label string
		Count int
	}

	buckets := make(map[string]*archiveBucket)
	order := make([]string, 0)

	for _, article := range articles {
		label := article.CreatedAt.Format("2006-01")
		if _, exists := buckets[label]; !exists {
			buckets[label] = &archiveBucket{
				Label: label,
				Count: 0,
			}
			order = append(order, label)
		}

		buckets[label].Count++
	}

	result := make([]map[string]any, 0, len(order))
	for _, label := range order {
		result = append(result, map[string]any{
			"label": buckets[label].Label,
			"count": buckets[label].Count,
		})
	}

	return result, nil
}

// ListComments 返回文章评论列表。
func (service *PortalService) ListComments(articleID int64) ([]model.Comment, error) {
	return service.commentRepo.ListByArticleID(articleID)
}

// CreateComment 创建前台评论。
func (service *PortalService) CreateComment(articleID int64, author string, content string) (model.Comment, error) {
	if _, err := service.articleRepo.FindByID(articleID); err != nil {
		return model.Comment{}, err
	}

	status := "approved"
	lowerContent := strings.ToLower(content)
	if strings.Contains(lowerContent, "spam") || strings.Contains(lowerContent, "广告") {
		status = "pending"
	}

	comment := model.Comment{
		ArticleID: articleID,
		Author:    strings.TrimSpace(author),
		Content:   strings.TrimSpace(content),
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if comment.Author == "" {
		comment.Author = "匿名读者"
	}

	if err := service.commentRepo.Create(&comment); err != nil {
		return model.Comment{}, err
	}

	return comment, nil
}
