package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/repository"
	"github.com/qingwa-ink/lychee/internal/service"
)

// CheckInController 打卡接口控制器。
type CheckInController struct {
	svc *service.CheckInService
}

func NewCheckInController(svc *service.CheckInService) *CheckInController {
	return &CheckInController{svc: svc}
}

type createRecordReq struct {
	Type       string  `json:"type" binding:"required"`
	Value      float64 `json:"value" binding:"required"`
	Unit       string  `json:"unit"`
	RecordDate string  `json:"record_date"` // 可空，默认今天
}

type upsertGoalReq struct {
	Type        string  `json:"type" binding:"required"`
	DailyTarget float64 `json:"daily_target" binding:"required"`
	Unit        string  `json:"unit"`
}

// ListRecords 查询打卡记录（query: date, type）。
func (ctrl *CheckInController) ListRecords(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	page, pageSize := parsePagination(c)
	list, total, err := ctrl.svc.ListRecords(c.Request.Context(), userID, repository.RecordFilter{
		Type:       c.Query("type"),
		RecordDate: c.Query("date"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// CreateRecord 新增打卡记录。
func (ctrl *CheckInController) CreateRecord(c *gin.Context) {
	var req createRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	rec, err := ctrl.svc.CreateRecord(c.Request.Context(), c.GetUint(CtxUserID), service.CreateRecordInput{
		Type: req.Type, Value: req.Value, Unit: req.Unit, RecordDate: req.RecordDate,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rec)
}

// Report 每日报告（query: date）。
func (ctrl *CheckInController) Report(c *gin.Context) {
	report, err := ctrl.svc.DailyReport(c.Request.Context(), c.GetUint(CtxUserID), c.Query("date"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, report)
}

// ListGoals 获取全部每日目标。
func (ctrl *CheckInController) ListGoals(c *gin.Context) {
	list, err := ctrl.svc.ListGoals(c.Request.Context(), c.GetUint(CtxUserID))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list})
}

// UpsertGoal 设置/更新某类型每日目标。
func (ctrl *CheckInController) UpsertGoal(c *gin.Context) {
	var req upsertGoalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	g, err := ctrl.svc.UpsertGoal(c.Request.Context(), c.GetUint(CtxUserID), service.UpsertGoalInput{
		Type: req.Type, DailyTarget: req.DailyTarget, Unit: req.Unit,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, g)
}
