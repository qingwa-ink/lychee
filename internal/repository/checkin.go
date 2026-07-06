package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// RecordFilter 打卡记录筛选参数。
type RecordFilter struct {
	Type       string
	RecordDate string // YYYY-MM-DD，为空则不限
	Page       int
	PageSize   int
}

// TypeSum 某类型在某日期的累计值（每日报告聚合用）。
type TypeSum struct {
	Type  string  `json:"type"`
	Total float64 `json:"total"`
}

// CheckInRecordRepository 打卡记录数据访问，全部按 user_id 隔离。
type CheckInRecordRepository struct {
	db *gorm.DB
}

func NewCheckInRecordRepository(db *gorm.DB) *CheckInRecordRepository {
	return &CheckInRecordRepository{db: db}
}

// Create 新增一条打卡记录。
func (r *CheckInRecordRepository) Create(rec *model.CheckInRecord) error {
	return r.db.Create(rec).Error
}

// List 按类型/日期分页查询记录，按创建时间倒序。count 与 find 各自从 r.db 构建。
func (r *CheckInRecordRepository) List(userID uint, f RecordFilter) ([]model.CheckInRecord, int64, error) {
	apply := func(tx *gorm.DB) *gorm.DB {
		tx = tx.Where("user_id = ?", userID)
		if f.Type != "" {
			tx = tx.Where("type = ?", f.Type)
		}
		if f.RecordDate != "" {
			tx = tx.Where("record_date = ?", f.RecordDate)
		}
		return tx
	}
	var total int64
	if err := apply(r.db).Model(&model.CheckInRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CheckInRecord
	if err := apply(r.db).Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// SumsByDate 按类型聚合某日期的累计值。
func (r *CheckInRecordRepository) SumsByDate(userID uint, date string) ([]TypeSum, error) {
	var sums []TypeSum
	err := r.db.Model(&model.CheckInRecord{}).
		Select("type, SUM(value) AS total").
		Where("user_id = ? AND record_date = ?", userID, date).
		Group("type").
		Scan(&sums).Error
	return sums, err
}

// CheckInGoalRepository 每日目标数据访问，(user_id, type) 唯一。
type CheckInGoalRepository struct {
	db *gorm.DB
}

func NewCheckInGoalRepository(db *gorm.DB) *CheckInGoalRepository {
	return &CheckInGoalRepository{db: db}
}

// ListByUser 查询某用户的全部每日目标。
func (r *CheckInGoalRepository) ListByUser(userID uint) ([]model.CheckInGoal, error) {
	var list []model.CheckInGoal
	if err := r.db.Where("user_id = ?", userID).Order("type ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByUserType 查询某用户某类型的目标（用于 upsert 判定）。
func (r *CheckInGoalRepository) FindByUserType(userID uint, typ string) (*model.CheckInGoal, error) {
	var g model.CheckInGoal
	if err := r.db.Where("user_id = ? AND type = ?", userID, typ).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// Upsert 新增或更新某用户某类型的每日目标。
func (r *CheckInGoalRepository) Upsert(g *model.CheckInGoal) error {
	var existing model.CheckInGoal
	err := r.db.Where("user_id = ? AND type = ?", g.UserID, g.Type).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]interface{}{
			"daily_target": g.DailyTarget,
			"unit":         g.Unit,
		}).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(g).Error
	}
	return err
}
