package service

import (
	"errors"
	"strings"

	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AdminService 负责后台管理相关业务。
type AdminService struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
	commentRepo  repository.CommentRepository
	userRepo     repository.UserRepository
}

// NewAdminService 创建后台服务。
func NewAdminService(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	commentRepo repository.CommentRepository,
	userRepo repository.UserRepository,
) *AdminService {
	return &AdminService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		commentRepo:  commentRepo,
		userRepo:     userRepo,
	}
}

// Dashboard 返回后台首页统计数据。
func (service *AdminService) Dashboard() (map[string]any, error) {
	articleResult, err := service.articleRepo.ListAdmin("", nil, 1, 1000)
	if err != nil {
		return nil, err
	}

	commentResult, err := service.commentRepo.ListAdmin("", "", 1, 1000)
	if err != nil {
		return nil, err
	}

	userResult, err := service.userRepo.List("", "", 1, 1000)
	if err != nil {
		return nil, err
	}

	publishedCount := 0
	totalViews := int64(0)
	for _, article := range articleResult.Items {
		totalViews += article.ViewCount
		if article.Status == 1 {
			publishedCount++
		}
	}

	categories, _ := service.categoryRepo.List()
	tags, _ := service.tagRepo.List()

	return map[string]any{
		"article_count":   articleResult.Total,
		"published_count": publishedCount,
		"view_count":      totalViews,
		"category_count":  len(categories),
		"tag_count":       len(tags),
		"comment_count":   commentResult.Total,
		"user_count":      userResult.Total,
	}, nil
}

// ListArticles 返回后台文章列表。
func (service *AdminService) ListArticles(keyword string, status *int, page int, pageSize int) (repository.ArticleListResult, error) {
	return service.articleRepo.ListAdmin(keyword, status, page, pageSize)
}

// SaveArticle 保存文章。
// 这里同时负责创建分类、创建标签，并把它们和文章关联起来。
func (service *AdminService) SaveArticle(
	id int64,
	title string,
	content string,
	summary string,
	coverURL string,
	status int,
	categoryName string,
	tagNames []string,
) (model.Article, error) {
	var categoryID *int64

	if strings.TrimSpace(categoryName) != "" {
		category, err := service.categoryRepo.FirstOrCreate(strings.TrimSpace(categoryName))
		if err != nil {
			return model.Article{}, err
		}
		categoryID = &category.ID
	}

	tags := make([]model.Tag, 0, len(tagNames))
	for _, tagName := range tagNames {
		trimmedName := strings.TrimSpace(tagName)
		if trimmedName == "" {
			continue
		}

		tag, err := service.tagRepo.FirstOrCreate(trimmedName)
		if err != nil {
			return model.Article{}, err
		}

		tags = append(tags, tag)
	}

	if id > 0 {
		article, err := service.articleRepo.FindByID(id)
		if err != nil {
			return model.Article{}, err
		}

		article.Title = title
		article.Content = content
		article.Summary = summary
		article.CoverURL = coverURL
		article.Status = status
		article.CategoryID = categoryID
		article.Tags = tags

		if err := service.articleRepo.Update(&article); err != nil {
			return model.Article{}, err
		}

		return service.articleRepo.FindByID(article.ID)
	}

	article := model.Article{
		Title:      title,
		Content:    content,
		Summary:    summary,
		CoverURL:   coverURL,
		Status:     status,
		CategoryID: categoryID,
		Tags:       tags,
		SourceType: "manual",
	}

	if err := service.articleRepo.Create(&article); err != nil {
		return model.Article{}, err
	}

	return service.articleRepo.FindByID(article.ID)
}

// DeleteArticle 删除文章。
func (service *AdminService) DeleteArticle(id int64) error {
	return service.articleRepo.Delete(id)
}

// ListCategories 返回分类列表。
func (service *AdminService) ListCategories() ([]model.Category, error) {
	return service.categoryRepo.List()
}

// CreateCategory 创建分类。
func (service *AdminService) CreateCategory(name string) (model.Category, error) {
	return service.categoryRepo.FirstOrCreate(strings.TrimSpace(name))
}

// ListTags 返回标签列表。
func (service *AdminService) ListTags() ([]model.Tag, error) {
	return service.tagRepo.List()
}

// CreateTag 创建标签。
func (service *AdminService) CreateTag(name string) (model.Tag, error) {
	return service.tagRepo.FirstOrCreate(strings.TrimSpace(name))
}

// ListComments 返回评论分页列表。
func (service *AdminService) ListComments(keyword string, status string, page int, pageSize int) (repository.CommentListResult, error) {
	return service.commentRepo.ListAdmin(keyword, status, page, pageSize)
}

// UpdateCommentStatus 修改评论状态。
func (service *AdminService) UpdateCommentStatus(commentID int64, status string) error {
	if status != "approved" && status != "pending" && status != "rejected" {
		return errors.New("评论状态不合法")
	}

	return service.commentRepo.UpdateStatus(commentID, status)
}

// DeleteComment 删除评论。
func (service *AdminService) DeleteComment(commentID int64) error {
	return service.commentRepo.Delete(commentID)
}

// ListUsers 返回用户分页列表。
func (service *AdminService) ListUsers(keyword string, role string, page int, pageSize int) (map[string]any, error) {
	result, err := service.userRepo.List(keyword, role, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, user := range result.Items {
		items = append(items, sanitizeUser(user))
	}

	return map[string]any{
		"items": items,
		"total": result.Total,
	}, nil
}

// CreateUser 创建后台用户。
func (service *AdminService) CreateUser(username string, password string, avatar string, role string) (map[string]any, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("用户名不能为空")
	}

	if len(strings.TrimSpace(password)) < 6 {
		return nil, errors.New("密码至少需要 6 位")
	}

	if role != "admin" && role != "user" {
		role = "user"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: strings.TrimSpace(username),
		Password: string(hashedPassword),
		Avatar:   strings.TrimSpace(avatar),
		Role:     role,
	}

	if user.Avatar == "" {
		user.Avatar = "https://placehold.co/120x120?text=User"
	}

	if err := service.userRepo.Create(&user); err != nil {
		return nil, err
	}

	return sanitizeUser(user), nil
}

// UpdateUserRole 修改用户角色。
func (service *AdminService) UpdateUserRole(userID int64, role string) error {
	if role != "admin" && role != "user" {
		return errors.New("用户角色不合法")
	}

	return service.userRepo.UpdateRole(userID, role)
}

// DeleteUser 删除用户。
func (service *AdminService) DeleteUser(currentUserID int64, targetUserID int64) error {
	if currentUserID == targetUserID {
		return errors.New("不能删除当前登录账号")
	}

	return service.userRepo.Delete(targetUserID)
}

// ProviderOverview 返回当前 AI Provider 概览。
func (service *AdminService) ProviderOverview(registry *ProviderRegistry) map[string]any {
	return map[string]any{
		"providers": registry.StatusList(),
		"message":   "这里展示当前系统已经接入的模型厂商，以及它们是否具备可用配置。",
	}
}
