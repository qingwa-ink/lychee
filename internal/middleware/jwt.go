// Package middleware 实现全局中间件（鉴权、限流、日志、i18n 等）。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/jwt"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// JWT 鉴权中间件：校验 Access Token，并将 user_id、locale 注入 context。
func JWT(jwtMgr *jwt.Manager, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil || claims.Type != jwt.TypeAccess {
			response.Fail(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(claims.UserID)
		if err != nil || user.TokenVersion != claims.Ver {
			response.Fail(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("locale", user.Locale)
		c.Next()
	}
}
