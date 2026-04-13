package repository

import (
	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SystemConfigRepository interface {
	ListAll() ([]model.SystemConfig, error)
	ListPublic() ([]model.SystemConfig, error)
	SaveMany(items []model.SystemConfig) error
}

type GormSystemConfigRepository struct {
	db *gorm.DB
}

func NewGormSystemConfigRepository(db *gorm.DB) *GormSystemConfigRepository {
	return &GormSystemConfigRepository{db: db}
}

func (repo *GormSystemConfigRepository) ListAll() ([]model.SystemConfig, error) {
	var items []model.SystemConfig
	err := repo.db.Order("config_group asc, id asc").Find(&items).Error
	return items, err
}

func (repo *GormSystemConfigRepository) ListPublic() ([]model.SystemConfig, error) {
	var items []model.SystemConfig
	err := repo.db.Where("is_public = ?", true).Order("config_group asc, id asc").Find(&items).Error
	return items, err
}

func (repo *GormSystemConfigRepository) SaveMany(items []model.SystemConfig) error {
	if len(items) == 0 {
		return nil
	}

	return repo.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"config_group",
			"name",
			"value",
			"input_type",
			"description",
			"is_public",
			"updated_at",
		}),
	}).Create(&items).Error
}
