package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/repository"
	"github.com/qingwa-ink/lychee/internal/service"
)

// TaskController 任务接口控制器。
type TaskController struct {
	svc *service.TaskService
}

func NewTaskController(svc *service.TaskService) *TaskController {
	return &TaskController{svc: svc}
}

type createTaskReq struct {
	GroupID  uint       `json:"group_id" binding:"required"`
	Content  string     `json:"content" binding:"required,max=5000"`
	Priority *int       `json:"priority"`
	Status   string     `json:"status"`
	DueDate  *time.Time `json:"due_date"`
}

type updateTaskReq struct {
	Content  *string    `json:"content"`
	Priority *int       `json:"priority"`
	Status   *string    `json:"status"`
	DueDate  *time.Time `json:"due_date"`
}

// List 任务列表（支持 group_id/status/priority 筛选与排序）。
func (ctrl *TaskController) List(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	page, pageSize := parsePagination(c)

	f := repository.TaskFilter{
		GroupID:  parseUintPtr(c.Query("group_id")),
		Status:   c.Query("status"),
		Priority: parseIntPtr(c.Query("priority")),
		Sort:     c.Query("sort"),
		Order:    c.DefaultQuery("order", "desc"),
		Page:     page,
		PageSize: pageSize,
	}
	list, total, err := ctrl.svc.List(c.Request.Context(), userID, f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// Create 新增任务。
func (ctrl *TaskController) Create(c *gin.Context) {
	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	t, err := ctrl.svc.Create(c.Request.Context(), c.GetUint(CtxUserID), service.CreateTaskInput{
		GroupID: req.GroupID, Content: req.Content, Priority: req.Priority, Status: req.Status, DueDate: req.DueDate,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

// Get 任务详情（前端可据此复制内容到剪贴板）。
func (ctrl *TaskController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	t, err := ctrl.svc.Get(c.Request.Context(), c.GetUint(CtxUserID), uint(id))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

// Update 修改任务（内容、优先级、状态、截止日期，均可选）。
func (ctrl *TaskController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	var req updateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	t, err := ctrl.svc.Update(c.Request.Context(), c.GetUint(CtxUserID), uint(id), service.UpdateTaskInput{
		Content: req.Content, Priority: req.Priority, Status: req.Status, DueDate: req.DueDate,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

// Delete 删除任务。
func (ctrl *TaskController) Delete(c *gin.Context) {
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

// parseUintPtr 解析可选 uint 查询参数。
func parseUintPtr(s string) *uint {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}
	uv := uint(v)
	return &uv
}

// parseIntPtr 解析可选 int 查询参数。
func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
