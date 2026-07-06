package controller

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/repository"
	"github.com/qingwa-ink/lychee/internal/service"
)

// LogController 日志接口控制器。
type LogController struct {
	svc *service.LogService
}

func NewLogController(svc *service.LogService) *LogController {
	return &LogController{svc: svc}
}

// Operations 操作历史（分页、时间范围、类别筛选）。
func (ctrl *LogController) Operations(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	page, pageSize := parsePagination(c)
	list, total, err := ctrl.svc.List(c.Request.Context(), userID, repository.LogFilter{
		Category: c.Query("category"),
		Start:    parseTimeQuery(c.Query("start")),
		End:      parseTimeQuery(c.Query("end")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// OperationsReport 操作报告：每日/每小时次数（柱状图数据）。
func (ctrl *LogController) OperationsReport(c *gin.Context) {
	buckets, err := ctrl.svc.Report(c.Request.Context(), c.GetUint(CtxUserID),
		c.DefaultQuery("dimension", "day"), c.Query("start"), c.Query("end"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"buckets": buckets})
}

// Logins 登录日志。
func (ctrl *LogController) Logins(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	page, pageSize := parsePagination(c)
	list, total, err := ctrl.svc.ListLogins(c.Request.Context(), userID, repository.LogFilter{
		Start:    parseTimeQuery(c.Query("start")),
		End:      parseTimeQuery(c.Query("end")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// parseTimeQuery 解析可选时间查询参数（RFC3339 或 YYYY-MM-DD）。
func parseTimeQuery(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}
