package repository

import (
	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

// TagRepository 定义标签仓库接口。
type TagRepository interface {
	List() ([]model.Tag, error)
	FirstOrCreate(name string) (model.Tag, error)
}

// GormTagRepository 是标签仓库的 GORM 实现。
type GormTagRepository struct {
	db *gorm.DB
}

// NewGormTagRepository 创建标签仓库。
func NewGormTagRepository(db *gorm.DB) *GormTagRepository {
	return &GormTagRepository{db: db}
}

// List 返回全部标签。
func (repo *GormTagRepository) List() ([]model.Tag, error) {
	var tags []model.Tag
	err := repo.db.Order("id desc").Find(&tags).Error
	return tags, err
}

// FirstOrCreate 确保标签存在。
func (repo *GormTagRepository) FirstOrCreate(name string) (model.Tag, error) {
	tag := model.Tag{Name: name}
	err := repo.db.Where("name = ?", name).FirstOrCreate(&tag).Error
	return tag, err
}
