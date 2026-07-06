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
		if !tryAuth(c, jwtMgr, userRepo) {
			response.Fail(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// JWTOptional 可选鉴权：携带有效 Token 则注入 user_id/locale，否则放行（匿名）。
func JWTOptional(jwtMgr *jwt.Manager, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tryAuth(c, jwtMgr, userRepo)
		c.Next()
	}
}

// tryAuth 尝试从 Authorization 头解析并校验用户；成功返回 true 并写入 context。
func tryAuth(c *gin.Context, jwtMgr *jwt.Manager, userRepo *repository.UserRepository) bool {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	tokenStr := strings.TrimPrefix(header, "Bearer ")

	claims, err := jwtMgr.Parse(tokenStr)
	if err != nil || claims.Type != jwt.TypeAccess {
		return false
	}
	user, err := userRepo.FindByID(claims.UserID)
	if err != nil || user.TokenVersion != claims.Ver {
		return false
	}

	c.Set("user_id", user.ID)
	c.Set("locale", user.Locale)
	return true
}
