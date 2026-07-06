package service

import (
	"context"
	"strings"
	"time"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// 任务状态
const (
	TaskStatusEditing   = "editing"
	TaskStatusPending   = "pending"
	TaskStatusCompleted = "completed"
)

// 优先级范围
const (
	PriorityMin = 0
	PriorityMax = 5
	PriorityDef = 3
)

// 允许排序的字段白名单（防止 SQL 注入）。
var allowedSortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"due_date":   "due_date",
	"priority":   "priority",
}

// CreateTaskInput 创建任务入参。
type CreateTaskInput struct {
	GroupID  uint
	Content  string
	Priority *int
	Status   string
	DueDate  *time.Time
}

// UpdateTaskInput 更新任务入参（指针为 nil 表示不更新）。
type UpdateTaskInput struct {
	Content  *string
	Priority *int
	Status   *string
	DueDate  *time.Time
}

// TaskService 任务业务。
type TaskService struct {
	taskRepo  *repository.TaskRepository
	groupRepo *repository.TaskGroupRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, groupRepo *repository.TaskGroupRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, groupRepo: groupRepo}
}

// List 查询任务列表（支持筛选、排序、分页）。
func (s *TaskService) List(ctx context.Context, userID uint, f repository.TaskFilter) ([]model.Task, int64, error) {
	// 校验排序字段与方向
	if col, ok := allowedSortFields[f.Sort]; ok {
		f.Sort = col
	} else {
		f.Sort = "created_at"
	}
	order := strings.ToLower(f.Order)
	if order != "asc" {
		order = "desc"
	}
	f.Order = order

	return s.taskRepo.List(userID, f)
}

// Create 新增任务，校验分组归属、优先级、状态。
func (s *TaskService) Create(ctx context.Context, userID uint, in CreateTaskInput) (*model.Task, error) {
	if _, err := s.groupRepo.FindByIDAndUser(in.GroupID, userID); err != nil {
		return nil, bizerr.New(bizerr.CodeBadRequest, "分组不存在")
	}
	priority := PriorityDef
	if in.Priority != nil {
		if !isValidPriority(*in.Priority) {
			return nil, bizerr.New(bizerr.CodeBadRequest, "优先级需在 0-5 之间")
		}
		priority = *in.Priority
	}
	status := TaskStatusEditing
	if in.Status != "" {
		if !isValidStatus(in.Status) {
			return nil, bizerr.New(bizerr.CodeBadRequest, "状态非法")
		}
		status = in.Status
	}
	t := &model.Task{
		UserID:   userID,
		GroupID:  in.GroupID,
		Content:  in.Content,
		Priority: priority,
		Status:   status,
		DueDate:  in.DueDate,
	}
	if err := s.taskRepo.Create(t); err != nil {
		return nil, bizerr.ErrInternal
	}
	return t, nil
}

// Update 部分更新任务字段。
func (s *TaskService) Update(ctx context.Context, userID, id uint, in UpdateTaskInput) (*model.Task, error) {
	t, err := s.taskRepo.FindByIDAndUser(id, userID)
	if err != nil {
		return nil, bizerr.ErrNotFound
	}
	fields := make(map[string]interface{})
	if in.Content != nil {
		fields["content"] = *in.Content
	}
	if in.Priority != nil {
		if !isValidPriority(*in.Priority) {
			return nil, bizerr.New(bizerr.CodeBadRequest, "优先级需在 0-5 之间")
		}
		fields["priority"] = *in.Priority
	}
	if in.Status != nil {
		if !isValidStatus(*in.Status) {
			return nil, bizerr.New(bizerr.CodeBadRequest, "状态非法")
		}
		fields["status"] = *in.Status
	}
	if in.DueDate != nil {
		fields["due_date"] = *in.DueDate
	}
	if len(fields) == 0 {
		return t, nil
	}
	if err := s.taskRepo.UpdateFields(id, fields); err != nil {
		return nil, bizerr.ErrInternal
	}
	return s.taskRepo.FindByIDAndUser(id, userID)
}

// Get 查询单个任务详情。
func (s *TaskService) Get(ctx context.Context, userID, id uint) (*model.Task, error) {
	t, err := s.taskRepo.FindByIDAndUser(id, userID)
	if err != nil {
		return nil, bizerr.ErrNotFound
	}
	return t, nil
}

// Delete 删除任务。
func (s *TaskService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.taskRepo.FindByIDAndUser(id, userID); err != nil {
		return bizerr.ErrNotFound
	}
	if err := s.taskRepo.Delete(id); err != nil {
		return bizerr.ErrInternal
	}
	return nil
}

func isValidPriority(p int) bool { return p >= PriorityMin && p <= PriorityMax }

func isValidStatus(s string) bool {
	switch s {
	case TaskStatusEditing, TaskStatusPending, TaskStatusCompleted:
		return true
	}
	return false
}
