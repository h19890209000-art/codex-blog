package model

import "time"

type DailyBriefing struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`

	BriefingDate string `json:"briefing_date" gorm:"column:briefing_date;type:char(10);not null;index:idx_briefing_date_status_sort"`
	Title        string `json:"title" gorm:"type:varchar(255);not null"`
	Summary      string `json:"summary" gorm:"type:text"`

	SourceName string `json:"source_name" gorm:"column:source_name;type:varchar(120);default:''"`
	SourceURL  string `json:"source_url" gorm:"column:source_url;type:varchar(800);default:''"`
	SourceHash string `json:"source_hash" gorm:"column:source_hash;type:varchar(64);default:'';uniqueIndex:idx_briefing_unique"`
	SourceType string `json:"source_type" gorm:"column:source_type;type:varchar(20);default:'manual';index"`

	Status int `json:"status" gorm:"type:tinyint;default:1;index:idx_briefing_date_status_sort"`

	SortOrder         int        `json:"sort_order" gorm:"column:sort_order;default:0;index:idx_briefing_date_status_sort"`
	Region            string     `json:"region" gorm:"type:varchar(20);default:'global'"`
	Language          string     `json:"language" gorm:"type:varchar(10);default:'en'"`
	OriginFeed        string     `json:"origin_feed" gorm:"column:origin_feed;type:varchar(255);default:''"`
	SourcePublishedAt *time.Time `json:"source_published_at" gorm:"column:source_published_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DailyBriefing) TableName() string {
	return "daily_briefings"
}
