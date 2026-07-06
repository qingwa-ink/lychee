// Package service 实现核心业务逻辑。
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"

	bizerr "github.com/qingwa-ink/lychee/internal/pkg/errors"
	"github.com/qingwa-ink/lychee/internal/pkg/jwt"
	"github.com/qingwa-ink/lychee/internal/pkg/mail"
	"github.com/qingwa-ink/lychee/internal/pkg/password"
	"github.com/qingwa-ink/lychee/internal/model"
	"github.com/qingwa-ink/lychee/internal/repository"
)

// 验证码用途
const (
	CodeTypeRegister       = "register"
	CodeTypeForgot         = "forgot"
	CodeTypeChangePassword = "change_password"
)

// AuthService 认证业务。
type AuthService struct {
	userRepo    *repository.UserRepository
	codeRepo    *repository.VerificationRepository
	refreshRepo *repository.RefreshTokenRepository
	jwt         *jwt.Manager
	mailer      mail.Mailer
	codeTTL     time.Duration
}

// NewAuthService 构造 AuthService。codeTTL 为验证码有效期。
func NewAuthService(
	userRepo *repository.UserRepository,
	codeRepo *repository.VerificationRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtMgr *jwt.Manager,
	mailer mail.Mailer,
	codeTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		codeRepo:    codeRepo,
		refreshRepo: refreshRepo,
		jwt:         jwtMgr,
		mailer:      mailer,
		codeTTL:     codeTTL,
	}
}

// SendCode 生成并发送验证码。同一 email+type 60s 内只能发一次。
func (s *AuthService) SendCode(ctx context.Context, email, typ string) error {
	if recent, err := s.codeRepo.FindLatestCreatedWithin(email, typ, 60*time.Second); err == nil && recent != nil {
		return bizerr.New(bizerr.CodeRateLimited, "验证码已发送，请 60 秒后再试")
	}

	code := generateNumericCode(6)
	rec := &model.EmailVerificationCode{
		Email:     email,
		Code:      code,
		Type:      typ,
		ExpiredAt: time.Now().Add(s.codeTTL),
	}
	if err := s.codeRepo.Create(rec); err != nil {
		return bizerr.ErrInternal
	}

	if err := s.mailer.Send(email, "荔枝小秘书 验证码", "您的验证码是："+code+"，有效期 10 分钟。"); err != nil {
		return bizerr.ErrInternal
	}
	return nil
}

// Register 校验验证码并创建用户。
func (s *AuthService) Register(ctx context.Context, email, plain, code string) error {
	rec, err := s.codeRepo.FindValidCode(email, CodeTypeRegister, code)
	if err != nil {
		return bizerr.New(bizerr.CodeBadRequest, "验证码无效或已过期")
	}
	if existing, err := s.userRepo.FindByEmail(email); err == nil && existing != nil {
		return bizerr.New(bizerr.CodeConflict, "该邮箱已注册")
	}
	hash, err := password.Hash(plain)
	if err != nil {
		return bizerr.ErrInternal
	}
	user := &model.User{Email: email, PasswordHash: hash, Locale: "zh", Status: 1}
	if err := s.userRepo.Create(user); err != nil {
		return bizerr.New(bizerr.CodeConflict, "该邮箱已注册")
	}
	_ = s.codeRepo.MarkUsed(rec.ID)
	return nil
}

// Login 校验账号密码并签发双 Token。
func (s *AuthService) Login(ctx context.Context, email, plain string) (access, refresh string, err error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", bizerr.New(bizerr.CodeUnauthorized, "邮箱或密码错误")
	}
	if !password.Compare(user.PasswordHash, plain) {
		return "", "", bizerr.New(bizerr.CodeUnauthorized, "邮箱或密码错误")
	}
	return s.issueTokens(user.ID, user.TokenVersion)
}

// Refresh 用 Refresh Token 换取新的双 Token（一次性轮换：作废旧 token、签发新 token）。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil || claims.Type != jwt.TypeRefresh {
		return "", "", bizerr.ErrTokenExpired
	}
	// 必须在库中且未撤销、未过期
	stored, err := s.refreshRepo.FindActiveByHash(hashToken(refreshToken))
	if err != nil || stored.UserID != claims.UserID {
		return "", "", bizerr.ErrTokenExpired
	}
	// 校验用户仍处于同一 token 版本（未被登出吊销）
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil || user.TokenVersion != claims.Ver {
		return "", "", bizerr.ErrTokenExpired
	}
	// 轮换：作废当前 refresh token，再签发新的
	_ = s.refreshRepo.Revoke(stored.ID)
	return s.issueTokens(user.ID, user.TokenVersion)
}

// Logout 自增 token 版本号并撤销全部 Refresh Token，使历史凭证立即失效。
func (s *AuthService) Logout(ctx context.Context, userID uint) error {
	if err := s.userRepo.IncrementTokenVersion(userID); err != nil {
		return bizerr.ErrInternal
	}
	_ = s.refreshRepo.RevokeAllByUser(userID)
	return nil
}

// Profile 返回当前用户信息。
func (s *AuthService) Profile(userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, bizerr.ErrNotFound
	}
	return user, nil
}

// issueTokens 签发 Access + Refresh 双 Token，并将 refresh token 的哈希落库。
func (s *AuthService) issueTokens(userID, ver uint) (access, refresh string, err error) {
	access, refresh, err = s.jwt.Issue(userID, ver)
	if err != nil {
		return "", "", bizerr.ErrInternal
	}
	rt := &model.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(refresh),
		ExpiresAt: time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err := s.refreshRepo.Create(rt); err != nil {
		return "", "", bizerr.ErrInternal
	}
	return access, refresh, nil
}

// hashToken 对 token 做 sha256，避免明文入库。
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// generateNumericCode 使用 crypto/rand 生成 n 位数字验证码。
func generateNumericCode(n int) string {
	const digits = "0123456789"
	max := big.NewInt(int64(len(digits)))
	out := make([]byte, n)
	for i := range out {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = digits[0] // 极小概率失败时退回 0
			continue
		}
		out[i] = digits[num.Int64()]
	}
	return string(out)
}
