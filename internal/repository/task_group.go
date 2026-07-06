package repository

import (
	"gorm.io/gorm"

	"github.com/qingwa-ink/lychee/internal/model"
)

// TaskGroupRepository 任务分组数据访问，全部按 user_id 隔离。
type TaskGroupRepository struct {
	db *gorm.DB
}

func NewTaskGroupRepository(db *gorm.DB) *TaskGroupRepository {
	return &TaskGroupRepository{db: db}
}

// Create 新增分组。
func (r *TaskGroupRepository) Create(g *model.TaskGroup) error {
	return r.db.Create(g).Error
}

// ListByUser 查询某用户全部分组（用于构建树）。
func (r *TaskGroupRepository) ListByUser(userID uint) ([]model.TaskGroup, error) {
	var list []model.TaskGroup
	if err := r.db.Where("user_id = ?", userID).
		Order("sort_order ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByIDAndUser 查询属于指定用户的分组（归属校验）。
func (r *TaskGroupRepository) FindByIDAndUser(id, userID uint) (*model.TaskGroup, error) {
	var g model.TaskGroup
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// UpdateFields 按字段映射更新分组（仅更新给出的字段）。
func (r *TaskGroupRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.TaskGroup{}).Where("id = ?", id).Updates(fields).Error
}

// FindChildIDs 查询指定父分组下的直接子分组 ID（用于级联删除的 BFS）。
func (r *TaskGroupRepository) FindChildIDs(userID uint, parentIDs []uint) ([]uint, error) {
	if len(parentIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	if err := r.db.Model(&model.TaskGroup{}).
		Where("user_id = ? AND parent_id IN ?", userID, parentIDs).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// DeleteGroupsAndTasks 在事务中级联软删除分组及其任务。
func (r *TaskGroupRepository) DeleteGroupsAndTasks(userID uint, groupIDs []uint) error {
	if len(groupIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND group_id IN ?", userID, groupIDs).
			Delete(&model.Task{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND id IN ?", userID, groupIDs).
			Delete(&model.TaskGroup{}).Error; err != nil {
			return err
		}
		return nil
	})
}
