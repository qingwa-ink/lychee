package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/render"
)

// PageController 负责服务端渲染页面（html/template 骨架，部分页注入 Jet 片段）。
type PageController struct {
	r *render.Renderer
}

func NewPageController(r *render.Renderer) *PageController {
	return &PageController{r: r}
}

// render 统一渲染：locale 由 i18n 中间件写入上下文。content 为内容模板名（如 page_login）。
func (ctrl *PageController) render(c *gin.Context, page, content, titleKey string, showNav bool, scripts []string) {
	ctrl.renderData(c, page, content, titleKey, showNav, scripts, nil)
}

// renderData 同 render，但可携带自定义数据（如 Jet 片段 HTML，由内容模板 {{.Data}} 透传）。
func (ctrl *PageController) renderData(c *gin.Context, page, content, titleKey string, showNav bool, scripts []string, data any) {
	locale := c.GetString("locale")
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := ctrl.r.Render(c.Writer, locale, content, &render.PageData{
		Page:    page,
		Title:   titleKey,
		ShowNav: showNav,
		Scripts: scripts,
		Data:    data,
	}); err != nil {
		c.String(500, "render error: %v", err)
	}
}

// 公开页（无导航）
func (ctrl *PageController) Login(c *gin.Context) {
	ctrl.render(c, "login", "page_login", "auth.login", false, []string{"/static/js/pages/login.js"})
}

func (ctrl *PageController) Register(c *gin.Context) {
	ctrl.render(c, "register", "page_register", "auth.register", false, []string{"/static/js/pages/register.js"})
}

func (ctrl *PageController) ForgotPassword(c *gin.Context) {
	ctrl.render(c, "forgot", "page_forgot", "auth.forgot_password", false, []string{"/static/js/pages/forgot.js"})
}

// 应用页（带导航，客户端鉴权）— 未实现页先用通用外壳占位
func (ctrl *PageController) appPage(page, titleKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctrl.render(c, page, "page_app", titleKey, true, []string{"/static/js/pages/app.js"})
	}
}

func (ctrl *PageController) AppCheckIn(c *gin.Context) {
	ctrl.render(c, "checkin", "page_checkin", "nav.checkin", true, []string{"/static/js/pages/checkin.js"})
}
func (ctrl *PageController) AppLogs(c *gin.Context)    { ctrl.appPage("logs", "nav.logs")(c) }

// /app/tasks — 任务看板（F1.3）：Jet 渲染两栏骨架，注入页面后由 tasks.js 接管交互。
func (ctrl *PageController) AppTasks(c *gin.Context) {
	locale := c.GetString("locale")
	board, err := ctrl.r.Fragment("task_board.jet", locale, nil)
	if err != nil {
		c.String(500, "render board: %v", err)
		return
	}
	ctrl.renderData(c, "tasks", "page_tasks", "nav.tasks", true, []string{"/static/js/pages/tasks.js"}, board)
}

// /app/phrases — 常用语管理（F1.2）
func (ctrl *PageController) AppPhrases(c *gin.Context) {
	ctrl.render(c, "phrases", "page_phrases", "nav.phrases", true, []string{"/static/js/pages/phrases.js"})
}

// /app/settings — 个人设置（改密、语种）（F1.2）
func (ctrl *PageController) AppSettings(c *gin.Context) {
	ctrl.render(c, "settings", "page_settings", "nav.settings", true, []string{"/static/js/pages/settings.js"})
}
