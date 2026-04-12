package repository

import (
	"strings"

	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

// CommentListResult 用来承接后台评论分页结果。
type CommentListResult struct {
	Items []model.Comment `json:"items"`
	Total int64           `json:"total"`
}

// CommentRepository 定义评论仓库接口。
type CommentRepository interface {
	FindByID(id int64) (model.Comment, error)
	ListByArticleID(articleID int64) ([]model.Comment, error)
	ListAdmin(keyword string, status string, page int, pageSize int) (CommentListResult, error)
	Create(comment *model.Comment) error
	UpdateStatus(id int64, status string) error
	Delete(id int64) error
}

// GormCommentRepository 是评论仓库的 GORM 实现。
type GormCommentRepository struct {
	db *gorm.DB
}

// NewGormCommentRepository 创建评论仓库。
func NewGormCommentRepository(db *gorm.DB) *GormCommentRepository {
	return &GormCommentRepository{db: db}
}

// FindByID 按评论 ID 查找评论。
func (repo *GormCommentRepository) FindByID(id int64) (model.Comment, error) {
	var comment model.Comment
	err := repo.db.Preload("Article").First(&comment, id).Error
	return comment, err
}

// ListByArticleID 返回某篇文章下面的已通过评论。
func (repo *GormCommentRepository) ListByArticleID(articleID int64) ([]model.Comment, error) {
	var comments []model.Comment
	err := repo.db.Where("article_id = ? AND status = ?", articleID, "approved").Order("id desc").Find(&comments).Error
	return comments, err
}

// ListAdmin 返回后台评论分页列表。
func (repo *GormCommentRepository) ListAdmin(keyword string, status string, page int, pageSize int) (CommentListResult, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	query := repo.db.Model(&model.Comment{}).Preload("Article")

	if strings.TrimSpace(keyword) != "" {
		likeKeyword := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("author LIKE ? OR content LIKE ?", likeKeyword, likeKeyword)
	}

	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return CommentListResult{}, err
	}

	var comments []model.Comment
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments).Error
	if err != nil {
		return CommentListResult{}, err
	}

	return CommentListResult{
		Items: comments,
		Total: total,
	}, nil
}

// Create 新建评论。
func (repo *GormCommentRepository) Create(comment *model.Comment) error {
	return repo.db.Create(comment).Error
}

// UpdateStatus 修改评论状态。
func (repo *GormCommentRepository) UpdateStatus(id int64, status string) error {
	return repo.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除评论。
func (repo *GormCommentRepository) Delete(id int64) error {
	return repo.db.Delete(&model.Comment{}, id).Error
}
