package repository

import (
	"strings"

	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

type DailyBriefingListResult struct {
	Items []model.DailyBriefing `json:"items"`
	Total int64                 `json:"total"`
}

type DailyBriefingRepository interface {
	ListPublicByDate(date string) ([]model.DailyBriefing, error)
	ListPublishedDates(limit int) ([]string, error)
	LatestPublishedDate() (string, error)
	ListAdmin(date string, keyword string, status *int, page int, pageSize int) (DailyBriefingListResult, error)
	FindByID(id int64) (model.DailyBriefing, error)
	Delete(id int64) error
	Create(briefing *model.DailyBriefing) error
	Update(briefing *model.DailyBriefing) error
	DeleteAutoByDate(date string) error
}

type GormDailyBriefingRepository struct {
	db *gorm.DB
}

func NewGormDailyBriefingRepository(db *gorm.DB) *GormDailyBriefingRepository {
	return &GormDailyBriefingRepository{db: db}
}

func (repo *GormDailyBriefingRepository) ListPublicByDate(date string) ([]model.DailyBriefing, error) {
	var items []model.DailyBriefing
	err := repo.db.
		Where("briefing_date = ? AND status = ?", strings.TrimSpace(date), 1).
		Order("sort_order asc, source_published_at desc, id asc").
		Find(&items).
		Error
	return items, err
}

func (repo *GormDailyBriefingRepository) ListPublishedDates(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 30
	}

	var dates []string
	err := repo.db.
		Model(&model.DailyBriefing{}).
		Where("status = ?", 1).
		Distinct("briefing_date").
		Order("briefing_date desc").
		Limit(limit).
		Pluck("briefing_date", &dates).
		Error
	return dates, err
}

func (repo *GormDailyBriefingRepository) LatestPublishedDate() (string, error) {
	var dates []string
	err := repo.db.
		Model(&model.DailyBriefing{}).
		Where("status = ?", 1).
		Distinct("briefing_date").
		Order("briefing_date desc").
		Limit(1).
		Pluck("briefing_date", &dates).
		Error
	if err != nil || len(dates) == 0 {
		return "", err
	}
	return dates[0], nil
}

func (repo *GormDailyBriefingRepository) ListAdmin(date string, keyword string, status *int, page int, pageSize int) (DailyBriefingListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := repo.db.Model(&model.DailyBriefing{})

	if strings.TrimSpace(date) != "" {
		query = query.Where("briefing_date = ?", strings.TrimSpace(date))
	}
	if strings.TrimSpace(keyword) != "" {
		likeKeyword := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR source_name LIKE ?", likeKeyword, likeKeyword, likeKeyword)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return DailyBriefingListResult{}, err
	}

	var items []model.DailyBriefing
	err := query.
		Order("briefing_date desc, sort_order asc, source_published_at desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).
		Error
	if err != nil {
		return DailyBriefingListResult{}, err
	}

	return DailyBriefingListResult{Items: items, Total: total}, nil
}

func (repo *GormDailyBriefingRepository) FindByID(id int64) (model.DailyBriefing, error) {
	var item model.DailyBriefing
	err := repo.db.First(&item, id).Error
	return item, err
}

func (repo *GormDailyBriefingRepository) Delete(id int64) error {
	return repo.db.Delete(&model.DailyBriefing{}, id).Error
}

func (repo *GormDailyBriefingRepository) Create(briefing *model.DailyBriefing) error {
	return repo.db.Create(briefing).Error
}

func (repo *GormDailyBriefingRepository) Update(briefing *model.DailyBriefing) error {
	return repo.db.Save(briefing).Error
}

func (repo *GormDailyBriefingRepository) DeleteAutoByDate(date string) error {
	return repo.db.Where("briefing_date = ? AND source_type = ?", strings.TrimSpace(date), "auto").Delete(&model.DailyBriefing{}).Error
}
