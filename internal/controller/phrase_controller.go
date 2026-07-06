package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/service"
)

// parsePagination 解析分页参数并做边界约束（供各列表接口复用）。
func parsePagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return
}

// PhraseController 常用语接口控制器。
type PhraseController struct {
	svc *service.PhraseService
}

func NewPhraseController(svc *service.PhraseService) *PhraseController {
	return &PhraseController{svc: svc}
}

type phraseContentReq struct {
	Content string `json:"content" binding:"required,max=2000"`
}

// List 常用语列表（分页，按更新时间倒序）。
func (ctrl *PhraseController) List(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	page, pageSize := parsePagination(c)
	list, total, err := ctrl.svc.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// Create 新增常用语。
func (ctrl *PhraseController) Create(c *gin.Context) {
	var req phraseContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	p, err := ctrl.svc.Create(c.Request.Context(), c.GetUint(CtxUserID), req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}

// Update 修改常用语。
func (ctrl *PhraseController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	var req phraseContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	p, err := ctrl.svc.Update(c.Request.Context(), c.GetUint(CtxUserID), uint(id), req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}

// Delete 删除常用语。
func (ctrl *PhraseController) Delete(c *gin.Context) {
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
