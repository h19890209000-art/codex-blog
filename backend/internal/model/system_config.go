package model

import "time"

// SystemConfig stores configurable site copy and other system-level settings.
type SystemConfig struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`

	Key string `json:"key" gorm:"type:varchar(120);not null;uniqueIndex"`

	Group string `json:"group" gorm:"column:config_group;type:varchar(80);not null;default:'system';index"`

	Name string `json:"name" gorm:"type:varchar(120);not null"`

	Value string `json:"value" gorm:"type:text"`

	InputType string `json:"input_type" gorm:"column:input_type;type:varchar(20);not null;default:'text'"`

	Description string `json:"description" gorm:"type:varchar(255);default:''"`

	IsPublic bool `json:"is_public" gorm:"column:is_public;not null;default:true;index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
