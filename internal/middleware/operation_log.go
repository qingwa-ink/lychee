// Package middleware 提供 Gin 中间件。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/controller"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// OperationLog 自动记录请求的操作日志中间件。
// 挂在 /api/v1 组上：c.Next() 执行真正的处理器后，再落库一条日志。
// 此时 JWT 中间件已写入 user_id；登录接口由控制器在成功后写入 user_id。
// 为避免读取日志自身造成噪声与读放大，跳过 /logs 路径。
func OperationLog(repo *repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先放行，待处理器执行完毕再落库
		c.Next()

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/logs") {
			return
		}

		category := "operation"
		action := c.FullPath() // 路由模板，稳定可读
		if action == "" {
			action = path
		}
		// 登录/登出归为 login 类别
		if strings.HasPrefix(path, "/api/v1/auth/login") ||
			strings.HasPrefix(path, "/api/v1/auth/logout") {
			category = "login"
		}

		userID := c.GetUint(controller.CtxUserID)
		var uid *uint
		if userID > 0 {
			uid = &userID
		}

		ua := c.Request.UserAgent()
		if len(ua) > 500 {
			ua = ua[:500]
		}

		_ = repo.Create(&model.OperationLog{
			UserID:   uid,
			Category: category,
			Action:   action,
			Method:   c.Request.Method,
			Path:     path,
			IP:       c.ClientIP(),
			UA:       ua,
			Params:   c.Request.URL.RawQuery,
		})
	}
}
