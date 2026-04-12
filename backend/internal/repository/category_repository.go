package repository

import (
	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

// CategoryRepository 定义分类仓库接口。
type CategoryRepository interface {
	List() ([]model.Category, error)
	FindByName(name string) (model.Category, error)
	FirstOrCreate(name string) (model.Category, error)
}

// GormCategoryRepository 是分类仓库的 GORM 实现。
type GormCategoryRepository struct {
	db *gorm.DB
}

// NewGormCategoryRepository 创建分类仓库。
func NewGormCategoryRepository(db *gorm.DB) *GormCategoryRepository {
	return &GormCategoryRepository{db: db}
}

// List 返回全部分类。
func (repo *GormCategoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	err := repo.db.Order("id desc").Find(&categories).Error
	return categories, err
}

// FindByName 按名称查找分类。
func (repo *GormCategoryRepository) FindByName(name string) (model.Category, error) {
	var category model.Category
	err := repo.db.Where("name = ?", name).First(&category).Error
	return category, err
}

// FirstOrCreate 确保分类存在。
func (repo *GormCategoryRepository) FirstOrCreate(name string) (model.Category, error) {
	category := model.Category{Name: name}
	err := repo.db.Where("name = ?", name).FirstOrCreate(&category).Error
	return category, err
}
