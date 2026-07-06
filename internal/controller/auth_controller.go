// Package controller 为 HTTP 控制器（薄层），负责参数校验与响应组装。
package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/response"
	"github.com/qingwa-ink/lychee/internal/service"
)

// CtxUserID 为 gin.Context 中存储当前用户 ID 的键。
const CtxUserID = "user_id"

// AuthController 认证接口控制器。
type AuthController struct {
	svc *service.AuthService
}

func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

type sendCodeReq struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register forgot change_password"`
}

func (ctrl *AuthController) SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	if err := ctrl.svc.SendCode(c.Request.Context(), req.Email, req.Type); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "验证码已发送"})
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Code     string `json:"code" binding:"required,len=6"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	if err := ctrl.svc.Register(c.Request.Context(), req.Email, req.Password, req.Code); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "注册成功"})
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	access, refresh, err := ctrl.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest)
		return
	}
	access, refresh, err := ctrl.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	if err := ctrl.svc.Logout(c.Request.Context(), userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已登出"})
}

func (ctrl *AuthController) Profile(c *gin.Context) {
	userID := c.GetUint(CtxUserID)
	user, err := ctrl.svc.Profile(userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{
		"id":     user.ID,
		"email":  user.Email,
		"locale": user.Locale,
	})
}
