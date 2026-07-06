package service

import (
	"context"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// GroupNode 分组树节点。
type GroupNode struct {
	model.TaskGroup
	Children []*GroupNode `json:"children"`
}

// TaskGroupService 任务分组业务。
type TaskGroupService struct {
	repo *repository.TaskGroupRepository
}

func NewTaskGroupService(repo *repository.TaskGroupRepository) *TaskGroupService {
	return &TaskGroupService{repo: repo}
}

// GetTree 构建当前用户的分组树（含嵌套）。
func (s *TaskGroupService) GetTree(ctx context.Context, userID uint) ([]*GroupNode, error) {
	groups, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, bizerr.ErrInternal
	}

	nodes := make(map[uint]*GroupNode, len(groups))
	for i := range groups {
		nodes[groups[i].ID] = &GroupNode{TaskGroup: groups[i]}
	}

	var roots []*GroupNode
	for i := range groups {
		g := groups[i]
		node := nodes[g.ID]
		// 存在且属于本用户的父分组则挂到父节点下，否则视为根
		if g.ParentID != nil && *g.ParentID != 0 {
			if parent, ok := nodes[*g.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots, nil
}

// Create 新增分组。若指定 parent_id，需校验父分组归属当前用户。
func (s *TaskGroupService) Create(ctx context.Context, userID uint, parentID *uint, name string) (*model.TaskGroup, error) {
	if parentID != nil && *parentID != 0 {
		if _, err := s.repo.FindByIDAndUser(*parentID, userID); err != nil {
			return nil, bizerr.New(bizerr.CodeBadRequest, "父分组不存在")
		}
	}
	g := &model.TaskGroup{UserID: userID, ParentID: parentID, Name: name}
	if err := s.repo.Create(g); err != nil {
		return nil, bizerr.ErrInternal
	}
	return g, nil
}

// Update 修改分组名与/或排序（均为可选，nil 表示不更新）。
func (s *TaskGroupService) Update(ctx context.Context, userID, id uint, name *string, sortOrder *int) error {
	if _, err := s.repo.FindByIDAndUser(id, userID); err != nil {
		return bizerr.ErrNotFound
	}
	fields := make(map[string]interface{})
	if name != nil {
		fields["name"] = *name
	}
	if sortOrder != nil {
		fields["sort_order"] = *sortOrder
	}
	if len(fields) == 0 {
		return nil
	}
	if err := s.repo.UpdateFields(id, fields); err != nil {
		return bizerr.ErrInternal
	}
	return nil
}

// Delete 删除分组，并级联软删除其所有子孙分组与任务。
func (s *TaskGroupService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.repo.FindByIDAndUser(id, userID); err != nil {
		return bizerr.ErrNotFound
	}
	// BFS 收集全部后代分组 ID
	allIDs := []uint{id}
	pending := []uint{id}
	for len(pending) > 0 {
		children, err := s.repo.FindChildIDs(userID, pending)
		if err != nil {
			return bizerr.ErrInternal
		}
		allIDs = append(allIDs, children...)
		pending = children
	}
	if err := s.repo.DeleteGroupsAndTasks(userID, allIDs); err != nil {
		return bizerr.ErrInternal
	}
	return nil
}
