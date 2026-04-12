package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 表示一篇文章。
// 这张表既保存手工创建的文章，也保存从 OSS 同步过来的文章。
type Article struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`

	// Title 是文章标题。
	Title string `json:"title" gorm:"type:varchar(255);not null;index"`

	// Content 保存 Markdown 正文。
	Content string `json:"content" gorm:"type:longtext;not null"`

	// Summary 保存摘要。
	Summary string `json:"summary" gorm:"type:text"`

	// CoverURL 保存封面图地址。
	CoverURL string `json:"cover_url" gorm:"column:cover_url;type:varchar(500);default:''"`

	// Status 表示文章状态。
	// 0 = 草稿
	// 1 = 发布
	Status int `json:"status" gorm:"type:tinyint;default:0;comment:0草稿 1发布"`

	// ViewCount 表示阅读量。
	ViewCount int64 `json:"view_count" gorm:"column:view_count;default:0"`

	// CategoryID 指向文章所属分类。
	CategoryID *int64    `json:"category_id" gorm:"column:category_id;index"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`

	// Tags 保存文章标签。
	Tags []Tag `json:"tags,omitempty" gorm:"many2many:article_tags;"`

	// SourceType 用来标记文章来源。
	// manual 表示后台手工创建。
	// oss 表示从阿里云 OSS 同步而来。
	SourceType string `json:"source_type" gorm:"column:source_type;type:varchar(30);default:'manual';uniqueIndex:idx_article_source"`

	// SourceKey 是外部来源里的唯一标识。
	// 这次 OSS 同步里，我们使用 “分类名/文件名” 作为唯一标识。
	SourceKey string `json:"source_key" gorm:"column:source_key;type:varchar(255);default:'';uniqueIndex:idx_article_source"`

	// SourcePath 保存 OSS 里的原始对象路径，方便排查问题。
	SourcePath string `json:"source_path" gorm:"column:source_path;type:varchar(500);default:''"`

	// SourceHash 保存本次同步后的内容摘要。
	// 这样下次同步时就能快速判断内容是否真的变了。
	SourceHash string `json:"source_hash" gorm:"column:source_hash;type:varchar(64);default:''"`

	// LastSyncedAt 记录最后一次同步时间。
	LastSyncedAt *time.Time `json:"last_synced_at" gorm:"column:last_synced_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名。
func (Article) TableName() string {
	return "articles"
}
