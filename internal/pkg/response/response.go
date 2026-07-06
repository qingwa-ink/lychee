// Package response 提供统一响应封装（见 doc/项目设计文档.md §5.1）。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qingwa-ink/lychee/internal/pkg/errors"
)

// Body 统一响应体。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: errors.CodeSuccess, Message: "ok", Data: data})
}

// Fail 返回失败响应。业务错误以 HTTP 200 + code 形式返回，由前端依据 code 处理。
func Fail(c *gin.Context, err error) {
	if be, ok := err.(*errors.BusinessError); ok {
		c.JSON(http.StatusOK, Body{Code: be.Code, Message: be.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, Body{Code: errors.CodeInternal, Message: "服务器内部错误"})
}
