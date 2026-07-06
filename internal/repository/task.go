package repository

import (
	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// TaskFilter 任务列表筛选与排序参数。
type TaskFilter struct {
	GroupID  *uint
	Status   string
	Priority *int
	Sort     string // 已白名单校验
	Order    string // asc / desc
	Page     int
	PageSize int
}

// TaskRepository 任务数据访问，全部按 user_id 隔离。
type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create 新增任务。
func (r *TaskRepository) Create(t *model.Task) error {
	return r.db.Create(t).Error
}

// List 按筛选条件分页查询任务。注意：count 与 find 各自从 r.db 构建链，
// 避免 GORM v2 中复用同一 *gorm.DB 时 Count() 污染 ORDER BY。
func (r *TaskRepository) List(userID uint, f TaskFilter) ([]model.Task, int64, error) {
	apply := func(tx *gorm.DB) *gorm.DB {
		tx = tx.Where("user_id = ?", userID)
		if f.GroupID != nil {
			tx = tx.Where("group_id = ?", *f.GroupID)
		}
		if f.Status != "" {
			tx = tx.Where("status = ?", f.Status)
		}
		if f.Priority != nil {
			tx = tx.Where("priority = ?", *f.Priority)
		}
		return tx
	}

	var total int64
	if err := apply(r.db).Model(&model.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Task
	if err := apply(r.db).Order(f.Sort + " " + f.Order).
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindByIDAndUser 查询属于指定用户的任务（归属校验）。
func (r *TaskRepository) FindByIDAndUser(id, userID uint) (*model.Task, error) {
	var t model.Task
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateFields 按字段映射更新任务（仅更新给出的字段）。
func (r *TaskRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 软删除任务。
func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}
