package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/i18n"
)

// I18N 解析请求语种并注入 context。优先级：cookie(lang) > Accept-Language > 默认。
// 注意：若后续 JWT 中间件命中已登录用户，会用其保存的 locale 覆盖此处结果。
func I18N(store *i18n.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieVal, _ := c.Cookie("lang")
		loc := store.Resolve(cookieVal, c.GetHeader("Accept-Language"))
		c.Set("locale", loc)
		c.Next()
	}
}
