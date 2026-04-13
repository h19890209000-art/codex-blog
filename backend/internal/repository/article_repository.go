package repository

import (
	"errors"
	"strings"
	"time"

	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

// ArticleListResult 用来承接“文章列表 + 总数”这种分页结构。
type ArticleListResult struct {
	Items []model.Article `json:"items"`
	Total int64           `json:"total"`
}

// ArticleRepository 定义文章仓库需要提供的方法。
type ArticleRepository interface {
	ListPublished(keyword string) ([]model.Article, error)
	ListAdmin(keyword string, status *int, page int, pageSize int) (ArticleListResult, error)
	FindByID(id int64) (model.Article, error)
	FindPublishedNavigation(id int64) (model.Article, model.Article, error)
	Search(keyword string) ([]model.Article, error)
	FindBySourceKey(sourceType string, sourceKey string) (model.Article, error)
	ListBySourceType(sourceType string) ([]model.Article, error)
	Create(article *model.Article) error
	Update(article *model.Article) error
	Restore(id int64) error
	Delete(id int64) error
}

// GormArticleRepository 是文章仓库的 GORM 实现。
type GormArticleRepository struct {
	db *gorm.DB
}

// NewGormArticleRepository 创建文章仓库。
func NewGormArticleRepository(db *gorm.DB) *GormArticleRepository {
	return &GormArticleRepository{db: db}
}

func (repo *GormArticleRepository) publishedPreviewQuery() *gorm.DB {
	return repo.db.
		Select("id", "title", "summary", "cover_url", "status", "view_count", "category_id", "created_at", "updated_at").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", 1)
}

// ListPublished 返回前台已发布文章。
func (repo *GormArticleRepository) ListPublished(keyword string) ([]model.Article, error) {
	query := repo.publishedPreviewQuery()

	if strings.TrimSpace(keyword) != "" {
		likeKeyword := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", likeKeyword, likeKeyword)
	}

	var articles []model.Article
	err := query.Order("id desc").Find(&articles).Error
	return articles, err
}

func (repo *GormArticleRepository) FindPublishedNavigation(id int64) (model.Article, model.Article, error) {
	var previous model.Article
	var next model.Article

	prevErr := repo.publishedPreviewQuery().
		Where("id > ?", id).
		Order("id asc").
		First(&previous).
		Error
	if prevErr != nil && !errors.Is(prevErr, gorm.ErrRecordNotFound) {
		return model.Article{}, model.Article{}, prevErr
	}

	nextErr := repo.publishedPreviewQuery().
		Where("id < ?", id).
		Order("id desc").
		First(&next).
		Error
	if nextErr != nil && !errors.Is(nextErr, gorm.ErrRecordNotFound) {
		return model.Article{}, model.Article{}, nextErr
	}

	if errors.Is(prevErr, gorm.ErrRecordNotFound) {
		previous = model.Article{}
	}

	if errors.Is(nextErr, gorm.ErrRecordNotFound) {
		next = model.Article{}
	}

	return previous, next, nil
}

// ListAdmin 返回后台文章分页列表。
func (repo *GormArticleRepository) ListAdmin(keyword string, status *int, page int, pageSize int) (ArticleListResult, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	query := repo.db.Model(&model.Article{}).Preload("Category").Preload("Tags")

	if strings.TrimSpace(keyword) != "" {
		likeKeyword := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", likeKeyword, likeKeyword)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ArticleListResult{}, err
	}

	var articles []model.Article
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error
	if err != nil {
		return ArticleListResult{}, err
	}

	return ArticleListResult{
		Items: articles,
		Total: total,
	}, nil
}

// FindByID 按 ID 查询文章。
func (repo *GormArticleRepository) FindByID(id int64) (model.Article, error) {
	var article model.Article
	err := repo.db.Preload("Category").Preload("Tags").First(&article, id).Error
	return article, err
}

// Search 按标题和内容搜索已发布文章。
func (repo *GormArticleRepository) Search(keyword string) ([]model.Article, error) {
	return repo.ListPublished(keyword)
}

// FindBySourceKey 按外部来源唯一标识查询文章。
// 这里使用 Unscoped 是为了让被软删除的文章也能被同步任务找回来并恢复。
func (repo *GormArticleRepository) FindBySourceKey(sourceType string, sourceKey string) (model.Article, error) {
	var article model.Article
	err := repo.db.Unscoped().
		Preload("Category").
		Preload("Tags").
		Where("source_type = ? AND source_key = ?", sourceType, sourceKey).
		First(&article).
		Error
	return article, err
}

// ListBySourceType 返回某个来源下的所有文章。
// 这里同样使用 Unscoped，这样软删除的数据也能参与“缺失文件判定”。
func (repo *GormArticleRepository) ListBySourceType(sourceType string) ([]model.Article, error) {
	var articles []model.Article
	err := repo.db.Unscoped().
		Preload("Category").
		Preload("Tags").
		Where("source_type = ?", sourceType).
		Find(&articles).
		Error
	return articles, err
}

// Create 新建文章。
func (repo *GormArticleRepository) Create(article *model.Article) error {
	return repo.db.Create(article).Error
}

// Update 更新文章。
func (repo *GormArticleRepository) Update(article *model.Article) error {
	if err := repo.db.Model(article).Association("Tags").Replace(article.Tags); err != nil {
		return err
	}

	return repo.db.Omit("Tags").Session(&gorm.Session{FullSaveAssociations: true}).Updates(article).Error
}

// Restore 恢复一篇被软删除的文章。
func (repo *GormArticleRepository) Restore(id int64) error {
	now := time.Now()

	return repo.db.Unscoped().
		Model(&model.Article{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"deleted_at":     nil,
			"last_synced_at": &now,
		}).
		Error
}

// Delete 软删除文章。
// 这里不再删除 article_tags 关联表数据，这样后续恢复时标签还能保留下来。
func (repo *GormArticleRepository) Delete(id int64) error {
	return repo.db.Delete(&model.Article{}, id).Error
}
