package model

import "time"

// User 表示后台管理员或普通用户。
type User struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	Avatar    string    `json:"avatar" gorm:"type:varchar(255);default:''"`
	Role      string    `json:"role" gorm:"type:varchar(20);default:user;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
