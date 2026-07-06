package repository

import (
	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// PhraseRepository 常用语数据访问，全部按 user_id 隔离。
type PhraseRepository struct {
	db *gorm.DB
}

func NewPhraseRepository(db *gorm.DB) *PhraseRepository {
	return &PhraseRepository{db: db}
}

// Create 新增常用语。
func (r *PhraseRepository) Create(p *model.Phrase) error {
	return r.db.Create(p).Error
}

// ListByUser 分页查询某用户的常用语，按更新时间倒序。count 与 find 各自从 r.db 构建。
func (r *PhraseRepository) ListByUser(userID uint, page, pageSize int) ([]model.Phrase, int64, error) {
	var total int64
	if err := r.db.Where("user_id = ?", userID).Model(&model.Phrase{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Phrase
	if err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindByIDAndUser 查询属于指定用户的常用语（用于归属校验）。
func (r *PhraseRepository) FindByIDAndUser(id, userID uint) (*model.Phrase, error) {
	var p model.Phrase
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateContent 更新常用语内容。
func (r *PhraseRepository) UpdateContent(id uint, content string) error {
	return r.db.Model(&model.Phrase{}).Where("id = ?", id).Update("content", content).Error
}

// Delete 软删除常用语。
func (r *PhraseRepository) Delete(id uint) error {
	return r.db.Delete(&model.Phrase{}, id).Error
}
