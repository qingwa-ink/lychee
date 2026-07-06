package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/service"
)

// TaskGroupController 任务分组接口控制器。
type TaskGroupController struct {
	svc *service.TaskGroupService
}

func NewTaskGroupController(svc *service.TaskGroupService) *TaskGroupController {
	return &TaskGroupController{svc: svc}
}

// Tree 返回分组树（含嵌套）。
func (ctrl *TaskGroupController) Tree(c *gin.Context) {
	tree, err := ctrl.svc.GetTree(c.Request.Context(), c.GetUint(CtxUserID))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"tree": tree})
}

type createGroupReq struct {
	ParentID *uint  `json:"parent_id"`
	Name     string `json:"name" binding:"required,max=100"`
}

// Create 新增分组。
func (ctrl *TaskGroupController) Create(c *gin.Context) {
	var req createGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	g, err := ctrl.svc.Create(c.Request.Context(), c.GetUint(CtxUserID), req.ParentID, req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, g)
}

type updateGroupReq struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
}

// Update 修改分组名与/或排序。
func (ctrl *TaskGroupController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	var req updateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	if err := ctrl.svc.Update(c.Request.Context(), c.GetUint(CtxUserID), uint(id), req.Name, req.SortOrder); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已更新"})
}

// Delete 删除分组（级联软删除子孙分组与任务）。
func (ctrl *TaskGroupController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	if err := ctrl.svc.Delete(c.Request.Context(), c.GetUint(CtxUserID), uint(id)); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已删除"})
}
