package model

import "time"

// Comment 表示文章评论。
type Comment struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ArticleID int64     `json:"article_id" gorm:"column:article_id;index"`
	Author    string    `json:"author" gorm:"type:varchar(100);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:approved"`
	Article   Article   `json:"article,omitempty" gorm:"foreignKey:ArticleID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Comment) TableName() string {
	return "comments"
}
