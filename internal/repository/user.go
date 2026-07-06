package repository

import (
	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// UserRepository 用户数据访问。
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户。
func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// FindByEmail 按邮箱查询用户。
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查询用户。
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdatePassword 更新密码哈希。
func (r *UserRepository) UpdatePassword(id uint, hash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", hash).Error
}

// UpdateLocale 更新语言偏好。
func (r *UserRepository) UpdateLocale(id uint, locale string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("locale", locale).Error
}

// IncrementTokenVersion 自增 token 版本号，使历史 Token 全部失效（登出 / 吊销）。
func (r *UserRepository) IncrementTokenVersion(id uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}
