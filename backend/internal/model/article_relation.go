package model

// ArticleTag 是文章和标签的关联表。
type ArticleTag struct {
	ArticleID int64 `json:"article_id" gorm:"primaryKey;column:article_id"`
	TagID     int64 `json:"tag_id" gorm:"primaryKey;column:tag_id"`
}

// TableName 指定表名。
func (ArticleTag) TableName() string {
	return "article_tags"
}
