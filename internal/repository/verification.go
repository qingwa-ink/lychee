package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// VerificationRepository 邮箱验证码数据访问。
type VerificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// Create 新增一条验证码记录。
func (r *VerificationRepository) Create(c *model.EmailVerificationCode) error {
	return r.db.Create(c).Error
}

// FindValidCode 查询匹配且未使用、未过期的最新验证码。
func (r *VerificationRepository) FindValidCode(email, typ, code string) (*model.EmailVerificationCode, error) {
	var c model.EmailVerificationCode
	err := r.db.Where("email = ? AND type = ? AND code = ? AND used = ? AND expired_at > ?",
		email, typ, code, false, time.Now()).
		Order("id DESC").First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// MarkUsed 标记验证码已使用。
func (r *VerificationRepository) MarkUsed(id uint) error {
	return r.db.Model(&model.EmailVerificationCode{}).Where("id = ?", id).Update("used", true).Error
}

// FindLatestCreatedWithin 查询指定时间窗口内最新创建的记录（用于发送频率限制）。
func (r *VerificationRepository) FindLatestCreatedWithin(email, typ string, d time.Duration) (*model.EmailVerificationCode, error) {
	var c model.EmailVerificationCode
	since := time.Now().Add(-d)
	err := r.db.Where("email = ? AND type = ? AND created_at > ?", email, typ, since).
		Order("id DESC").First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}
