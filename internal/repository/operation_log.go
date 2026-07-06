package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// LogFilter 操作日志筛选参数。
type LogFilter struct {
	Category string // login / operation，为空则不限
	Start    *time.Time
	End      *time.Time
	Page     int
	PageSize int
}

// BucketCount 报告中一个时间桶的计数（柱状图数据点）。
type BucketCount struct {
	Bucket string `json:"bucket"` // day: 2026-07-06；hour: 2026-07-06 14:00
	Count  int64  `json:"count"`
}

// OperationLogRepository 操作日志数据访问，按 user_id 隔离。
type OperationLogRepository struct {
	db *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create 落库一条操作日志（最佳努力：调用方应忽略错误）。
func (r *OperationLogRepository) Create(l *model.OperationLog) error {
	return r.db.Create(l).Error
}

// List 按筛选条件分页查询，按时间倒序。count 与 find 各自从 r.db 构建。
func (r *OperationLogRepository) List(userID uint, f LogFilter) ([]model.OperationLog, int64, error) {
	apply := func(tx *gorm.DB) *gorm.DB {
		tx = tx.Where("user_id = ?", userID)
		if f.Category != "" {
			tx = tx.Where("category = ?", f.Category)
		}
		if f.Start != nil {
			tx = tx.Where("created_at >= ?", *f.Start)
		}
		if f.End != nil {
			tx = tx.Where("created_at < ?", *f.End)
		}
		return tx
	}
	var total int64
	if err := apply(r.db).Model(&model.OperationLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.OperationLog
	if err := apply(r.db).Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Report 按维度（day/hour）聚合某用户在时间范围内的操作次数。
// created_at 带时区偏移存储，DATE/strftime 默认按 UTC 输出，故加 'localtime'
// 使桶与服务器本地时区（用户视角）一致，避免午夜附近的日期错位。
func (r *OperationLogRepository) Report(userID uint, dimension string, start, end time.Time) ([]BucketCount, error) {
	expr := "DATE(created_at, 'localtime')"
	if dimension == "hour" {
		expr = "strftime('%Y-%m-%d %H:00', created_at, 'localtime')"
	}
	var out []BucketCount
	err := r.db.Model(&model.OperationLog{}).
		Select(expr+" AS bucket, COUNT(*) AS count").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Group("bucket").
		Order("bucket ASC").
		Scan(&out).Error
	return out, err
}
