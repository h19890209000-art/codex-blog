package repository

import (
	"strings"

	"ai-blog/backend/internal/model"

	"gorm.io/gorm"
)

// UserListResult 用来承接后台用户分页结果。
type UserListResult struct {
	Items []model.User `json:"items"`
	Total int64        `json:"total"`
}

// UserRepository 定义用户仓库接口。
type UserRepository interface {
	FindByUsername(username string) (model.User, error)
	FindByID(id int64) (model.User, error)
	UpdatePassword(userID int64, hashedPassword string) error
	List(keyword string, role string, page int, pageSize int) (UserListResult, error)
	Create(user *model.User) error
	UpdateRole(userID int64, role string) error
	Delete(userID int64) error
}

// GormUserRepository 是用户仓库的 GORM 实现。
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository 创建用户仓库。
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByUsername 按用户名查找用户。
func (repo *GormUserRepository) FindByUsername(username string) (model.User, error) {
	var user model.User
	err := repo.db.Where("username = ?", username).First(&user).Error
	return user, err
}

// FindByID 按 ID 查找用户。
func (repo *GormUserRepository) FindByID(id int64) (model.User, error) {
	var user model.User
	err := repo.db.First(&user, id).Error
	return user, err
}

// UpdatePassword 更新用户密码。
func (repo *GormUserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	return repo.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// List 返回后台用户分页列表。
func (repo *GormUserRepository) List(keyword string, role string, page int, pageSize int) (UserListResult, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	query := repo.db.Model(&model.User{})

	if strings.TrimSpace(keyword) != "" {
		likeKeyword := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("username LIKE ?", likeKeyword)
	}

	if strings.TrimSpace(role) != "" {
		query = query.Where("role = ?", strings.TrimSpace(role))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return UserListResult{}, err
	}

	var users []model.User
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	if err != nil {
		return UserListResult{}, err
	}

	return UserListResult{
		Items: users,
		Total: total,
	}, nil
}

// Create 新建用户。
func (repo *GormUserRepository) Create(user *model.User) error {
	return repo.db.Create(user).Error
}

// UpdateRole 修改用户角色。
func (repo *GormUserRepository) UpdateRole(userID int64, role string) error {
	return repo.db.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

// Delete 删除用户。
func (repo *GormUserRepository) Delete(userID int64) error {
	return repo.db.Delete(&model.User{}, userID).Error
}
