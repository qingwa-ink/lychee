// Package middleware 提供 Gin 中间件。
package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
)

// rateBucket 单个限流器的状态。
type rateBucket struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval time.Duration
}

// newRateBucket 构造限流器：同一 key 两次放行至少间隔 interval。
// perSecond<=0 视为不限流（返回 nil，调用方据此跳过中间件）。
func newRateBucket(perSecond int) *rateBucket {
	if perSecond <= 0 {
		return nil
	}
	return &rateBucket{
		last:     make(map[string]time.Time),
		interval: time.Second / time.Duration(perSecond),
	}
}

// allow 判断 key 是否允许放行，并更新最近放行时间。
// map 超过阈值时顺手清理过期项，避免内存无限增长。
func (b *rateBucket) allow(key string, now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if last, ok := b.last[key]; ok && now.Sub(last) < b.interval {
		return false
	}
	if len(b.last) > 100000 {
		for k, t := range b.last {
			if now.Sub(t) >= b.interval {
				delete(b.last, k)
			}
		}
	}
	b.last[key] = now
	return true
}

// RateLimit 每 IP 每「方法+路由」每秒至多 perSecond 次，超限返回 429（业务码 4290）。
// key 含 HTTP 方法：避免「写后立即读」（如 POST /phrases 紧跟 GET /phrases 刷新列表）
// 被误判为同一路由的连续请求而限流。perSecond<=0 时不启用限流（放行全部）。
func RateLimit(perSecond int) gin.HandlerFunc {
	bucket := newRateBucket(perSecond)
	return func(c *gin.Context) {
		if bucket == nil {
			c.Next()
			return
		}
		key := c.ClientIP() + "|" + c.Request.Method + "|" + c.FullPath()
		if !bucket.allow(key, time.Now()) {
			response.Fail(c, errors.ErrRateLimited) // HTTP 200 + code 4290
			c.Abort()
			return
		}
		c.Next()
	}
}
