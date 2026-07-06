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
	I18NMiddleware         gin.HandlerFunc
	JWTMiddleware          gin.HandlerFunc
	JWTOptionalMiddleware  gin.HandlerFunc
	AuthController         *controller.AuthController
	LocaleController       *controller.LocaleController
	PhraseController       *controller.PhraseController
	TaskGroupController    *controller.TaskGroupController
	TaskController         *controller.TaskController
	CheckInController      *controller.CheckInController
	LogController          *controller.LogController
	OperationLogMiddleware gin.HandlerFunc
	RateLimitMiddleware    gin.HandlerFunc
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

	// API v1：限流最先执行 → i18n 解析语种 → 操作日志落库
	api := r.Group("/api/v1")
	api.Use(deps.RateLimitMiddleware, deps.I18NMiddleware, deps.OperationLogMiddleware)

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

	// --- 需登录的业务接口 ---
	phrases := api.Group("/phrases")
	phrases.Use(deps.JWTMiddleware)
	{
		phrases.GET("", deps.PhraseController.List)
		phrases.POST("", deps.PhraseController.Create)
		phrases.PUT("/:id", deps.PhraseController.Update)
		phrases.DELETE("/:id", deps.PhraseController.Delete)
	}

	// 任务分组（嵌套树 + 级联删除）
	taskGroups := api.Group("/task-groups")
	taskGroups.Use(deps.JWTMiddleware)
	{
		taskGroups.GET("", deps.TaskGroupController.Tree)
		taskGroups.POST("", deps.TaskGroupController.Create)
		taskGroups.PUT("/:id", deps.TaskGroupController.Update)
		taskGroups.DELETE("/:id", deps.TaskGroupController.Delete)
	}

	// 任务（CRUD + 筛选排序）
	tasks := api.Group("/tasks")
	tasks.Use(deps.JWTMiddleware)
	{
		tasks.GET("", deps.TaskController.List)
		tasks.POST("", deps.TaskController.Create)
		tasks.GET("/:id", deps.TaskController.Get)
		tasks.PUT("/:id", deps.TaskController.Update)
		tasks.DELETE("/:id", deps.TaskController.Delete)
	}

	// 打卡与健康（记录/每日目标/每日报告）
	checkin := api.Group("/check-in")
	checkin.Use(deps.JWTMiddleware)
	{
		checkin.GET("/records", deps.CheckInController.ListRecords)
		checkin.POST("/records", deps.CheckInController.CreateRecord)
		checkin.GET("/report", deps.CheckInController.Report)
		checkin.GET("/goals", deps.CheckInController.ListGoals)
		checkin.PUT("/goals", deps.CheckInController.UpsertGoal)
	}

	// 操作日志与报告
	logs := api.Group("/logs")
	logs.Use(deps.JWTMiddleware)
	{
		logs.GET("/operations", deps.LogController.Operations)
		logs.GET("/operations/report", deps.LogController.OperationsReport)
		logs.GET("/logins", deps.LogController.Logins)
	}

	return r
}
