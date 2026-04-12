package bootstrap

import (
	"fmt"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDatabase 负责初始化数据库。
// 它会依次做 4 件事：
// 1. 先连接到 MySQL 服务本身。
// 2. 如果业务数据库不存在，就自动创建。
// 3. 连接到业务数据库并自动建表。
// 4. 写入默认管理员和演示数据。
func InitDatabase(cfg config.AppConfig) (*gorm.DB, error) {
	// 这里先构造一个“不指定数据库名”的连接。
	// 这样我们才能先执行 CREATE DATABASE。
	rootDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Charset,
	)

	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 如果业务数据库还不存在，这里会自动创建。
	createDatabaseSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s COLLATE %s_unicode_ci",
		cfg.Database.DatabaseName,
		cfg.Database.Charset,
		cfg.Database.Charset,
	)
	if err := rootDB.Exec(createDatabaseSQL).Error; err != nil {
		return nil, err
	}

	// 再构造真正连接业务数据库的 DSN。
	appDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DatabaseName,
		cfg.Database.Charset,
	)

	appDB, err := gorm.Open(mysql.Open(appDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate 会自动创建或补齐这些表结构。
	if err := appDB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleTag{},
		&model.Comment{},
		&model.DailyBriefing{},
	); err != nil {
		return nil, err
	}

	// 确保默认管理员存在。
	if err := seedDefaultAdmin(appDB, cfg); err != nil {
		return nil, err
	}

	// 确保第一次打开系统时，至少有一篇演示文章。
	if err := seedDefaultContent(appDB); err != nil {
		return nil, err
	}

	return appDB, nil
}

// seedDefaultAdmin 用来写入默认管理员。
// 如果管理员已经存在，就不会重复创建。
func seedDefaultAdmin(db *gorm.DB, cfg config.AppConfig) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", cfg.Auth.DefaultAdminUsername).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// 密码必须先加密，不能明文存数据库。
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Auth.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUser := model.User{
		Username: cfg.Auth.DefaultAdminUsername,
		Password: string(hashedPassword),
		Avatar:   "https://placehold.co/120x120?text=Admin",
		Role:     "admin",
	}

	return db.Create(&adminUser).Error
}

// seedDefaultContent 写入一份演示内容。
// 这样你第一次打开后台时，不会看到一个完全空白的系统。
func seedDefaultContent(db *gorm.DB) error {
	var articleCount int64
	if err := db.Model(&model.Article{}).Count(&articleCount).Error; err != nil {
		return err
	}

	if articleCount > 0 {
		return nil
	}

	goCategory := model.Category{Name: "Go"}
	aiTag := model.Tag{Name: "AI"}
	ginTag := model.Tag{Name: "Gin"}

	if err := db.FirstOrCreate(&goCategory, model.Category{Name: "Go"}).Error; err != nil {
		return err
	}

	if err := db.FirstOrCreate(&aiTag, model.Tag{Name: "AI"}).Error; err != nil {
		return err
	}

	if err := db.FirstOrCreate(&ginTag, model.Tag{Name: "Gin"}).Error; err != nil {
		return err
	}

	article := model.Article{
		Title:      "Go 新手如何理解 Gin 路由分层",
		Content:    "Gin 的 router 负责定义路由，controller 负责接收参数，service 负责业务处理，repository 负责读写数据库。这样分层后，你就能更快定位问题出在哪一层。",
		Summary:    "从 PHP 控制器思维迁移到 Go 的 router/controller/service 分层。",
		CoverURL:   "https://placehold.co/1200x630?text=Go+Gin",
		Status:     1,
		ViewCount:  128,
		CategoryID: &goCategory.ID,
		Tags:       []model.Tag{aiTag, ginTag},
		SourceType: "manual",
	}

	if err := db.Create(&article).Error; err != nil {
		return err
	}

	comment := model.Comment{
		ArticleID: article.ID,
		Author:    "新手读者",
		Content:   "原来 router 和 controller 的职责真的不一样，这里终于看明白了。",
		Status:    "approved",
	}

	return db.Create(&comment).Error
}
