// Package jwt 封装 Access / Refresh 双 Token 的签发与解析。
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token 类型
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

// Manager JWT 管理器。
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// Claims 自定义声明。
type Claims struct {
	UserID uint   `json:"uid"`
	Ver    uint   `json:"ver"` // 对应 user.token_version，用于服务端吊销
	Type   string `json:"typ"` // access / refresh
	jwt.RegisteredClaims
}

// NewManager 构造 JWT 管理器，TTL 以字符串（如 "15m"）传入。
func NewManager(secret, accessTTLStr, refreshTTLStr string) (*Manager, error) {
	at, err := time.ParseDuration(accessTTLStr)
	if err != nil {
		return nil, errors.New("invalid access_ttl: " + err.Error())
	}
	rt, err := time.ParseDuration(refreshTTLStr)
	if err != nil {
		return nil, errors.New("invalid refresh_ttl: " + err.Error())
	}
	return &Manager{secret: []byte(secret), accessTTL: at, refreshTTL: rt}, nil
}

// Issue 同时签发 Access 与 Refresh Token。每次签发生成唯一 jti，
// 即使同一秒内签发也不会产生相同 Token。
func (m *Manager) Issue(userID, ver uint) (access, refresh string, err error) {
	now := time.Now()
	access, err = m.sign(Claims{
		UserID: userID, Ver: ver, Type: TypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{ID: newJTI(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL))},
	})
	if err != nil {
		return
	}
	refresh, err = m.sign(Claims{
		UserID: userID, Ver: ver, Type: TypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{ID: newJTI(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL))},
	})
	return
}

// RefreshTTL 返回 Refresh Token 有效期，供落库 ExpiresAt 使用。
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) sign(c Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
}

// Parse 解析并校验 Token。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := t.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return c, nil
}

// newJTI 生成 16 字节随机 hex 作为唯一 Token ID。
func newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
