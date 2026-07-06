package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/i18n"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/service"
)

// LocaleController 多语言接口控制器。
type LocaleController struct {
	store *i18n.Store
	svc   *service.AuthService
}

func NewLocaleController(store *i18n.Store, svc *service.AuthService) *LocaleController {
	return &LocaleController{store: store, svc: svc}
}

// Get 返回当前语种与可用语种列表。
func (ctrl *LocaleController) Get(c *gin.Context) {
	response.OK(c, gin.H{
		"current":   c.GetString("locale"),
		"default":   ctrl.store.Default(),
		"available": ctrl.store.Languages(),
	})
}

type setLocaleReq struct {
	Locale string `json:"locale" binding:"required,oneof=zh en"`
}

// Set 切换语种：写入 Cookie；若已登录则同时持久化到用户偏好。
func (ctrl *LocaleController) Set(c *gin.Context) {
	var req setLocaleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	// 写入 Cookie（1 年），后续请求自动携带
	c.SetCookie("lang", req.Locale, 31536000, "/", "", false, true)

	// 已登录则持久化
	if userID := c.GetUint(CtxUserID); userID != 0 {
		_ = ctrl.svc.SetLocale(userID, req.Locale)
	}

	response.OK(c, gin.H{"current": req.Locale})
}

// Messages 返回某语种的全量文案（供前端渲染）。未指定时使用当前语种。
func (ctrl *LocaleController) Messages(c *gin.Context) {
	loc := c.Query("locale")
	if loc == "" || !ctrl.store.Exists(loc) {
		loc = c.GetString("locale")
	}
	response.OK(c, ctrl.store.Messages(loc))
}
