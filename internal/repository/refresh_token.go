package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// RefreshTokenRepository 管理 Refresh Token 的存储与状态。
type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create 落库一条 Refresh Token 记录（仅存哈希）。
func (r *RefreshTokenRepository) Create(rt *model.RefreshToken) error {
	return r.db.Create(rt).Error
}

// FindActiveByHash 按哈希查询有效（未撤销、未过期）的记录。
func (r *RefreshTokenRepository) FindActiveByHash(hash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now()).
		First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Revoke 标记单条记录为已撤销。
func (r *RefreshTokenRepository) Revoke(id uint) error {
	return r.db.Model(&model.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

// RevokeAllByUser 撤销某用户全部 Refresh Token（登出 / 吊销）。
func (r *RefreshTokenRepository) RevokeAllByUser(userID uint) error {
	return r.db.Model(&model.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
