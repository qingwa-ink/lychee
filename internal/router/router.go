// Package router 注册路由与中间件。
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/config"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
)

// New 构造 Gin 引擎。后续里程碑会在 /api/v1 下挂载业务路由。
func New(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// API v1 分组（业务路由在后续里程碑中挂载）
	_ = r.Group("/api/v1")

	return r
}
