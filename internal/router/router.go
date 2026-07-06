// Package router 注册路由与中间件。
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/config"
	"github.com/qingwa-ink/lychee/internal/controller"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
)

// Deps 路由所需依赖。各里程碑按需填充。
type Deps struct {
	I18NMiddleware        gin.HandlerFunc
	JWTMiddleware         gin.HandlerFunc
	JWTOptionalMiddleware gin.HandlerFunc
	AuthController        *controller.AuthController
	LocaleController      *controller.LocaleController
	// 后续里程碑：PhraseController / TaskController / ...
}

// New 构造 Gin 引擎并注册路由。
func New(cfg *config.Config, deps *Deps) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// API v1：统一挂载 i18n 中间件解析语种
	api := r.Group("/api/v1")
	api.Use(deps.I18NMiddleware)

	// --- 多语言（可选鉴权：匿名也能切换，登录则持久化） ---
	locale := api.Group("/locale")
	locale.Use(deps.JWTOptionalMiddleware)
	{
		locale.GET("", deps.LocaleController.Get)
		locale.PUT("", deps.LocaleController.Set)
		locale.GET("/messages", deps.LocaleController.Messages)
	}

	// --- 认证：公开接口 ---
	auth := api.Group("/auth")
	{
		auth.POST("/send-code", deps.AuthController.SendCode)
		auth.POST("/register", deps.AuthController.Register)
		auth.POST("/login", deps.AuthController.Login)
		auth.POST("/refresh", deps.AuthController.Refresh)
		auth.POST("/forgot-password", deps.AuthController.ForgotPassword)
	}

	// --- 认证：需登录 ---
	authAuthed := api.Group("/auth")
	authAuthed.Use(deps.JWTMiddleware)
	{
		authAuthed.POST("/logout", deps.AuthController.Logout)
		authAuthed.GET("/profile", deps.AuthController.Profile)
		authAuthed.PUT("/password", deps.AuthController.ChangePassword)
	}

	// --- 需登录的业务接口（后续里程碑挂载） ---
	authed := api.Group("/")
	authed.Use(deps.JWTMiddleware)

	return r
}
