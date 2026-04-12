package service

import (
	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"
)

// ArticleService 负责文章相关业务。
type ArticleService struct {
	articleRepo repository.ArticleRepository
}

// NewArticleService 创建文章服务。
func NewArticleService(articleRepo repository.ArticleRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
	}
}

// ListPublic 返回前台文章列表。
func (service *ArticleService) ListPublic(keyword string) ([]model.Article, error) {
	return service.articleRepo.ListPublished(keyword)
}

// Detail 返回文章详情。
func (service *ArticleService) Detail(id int64) (model.Article, error) {
	return service.articleRepo.FindByID(id)
}

// Search 返回文章搜索结果。
func (service *ArticleService) Search(keyword string) ([]model.Article, error) {
	return service.articleRepo.Search(keyword)
}
