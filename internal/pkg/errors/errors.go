// Package errors 定义统一错误码与业务错误类型（见 doc/项目设计文档.md §5.1）。
package errors

import "fmt"

// 错误码
const (
	CodeSuccess      = 0
	CodeBadRequest   = 4000
	CodeUnauthorized = 4010
	CodeTokenExpired = 4011 // Refresh Token 失效，需重新登录
	CodeForbidden    = 4030
	CodeNotFound     = 4040
	CodeConflict     = 4090
	CodeRateLimited  = 4290
	CodeInternal     = 5000
)

// BusinessError 携带错误码的业务错误。
type BusinessError struct {
	Code    int
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 构造一个业务错误。
func New(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Message: msg}
}

// 预定义错误
var (
	ErrBadRequest   = New(CodeBadRequest, "参数错误")
	ErrUnauthorized = New(CodeUnauthorized, "未认证或登录已过期")
	ErrTokenExpired = New(CodeTokenExpired, "凭证已失效，请重新登录")
	ErrForbidden    = New(CodeForbidden, "无权限访问该资源")
	ErrNotFound     = New(CodeNotFound, "资源不存在")
	ErrConflict     = New(CodeConflict, "资源冲突")
	ErrRateLimited  = New(CodeRateLimited, "请求过于频繁，请稍后再试")
	ErrInternal     = New(CodeInternal, "服务器内部错误")
)
