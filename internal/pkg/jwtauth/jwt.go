// Package jwtauth 负责后台管理 JWT 的签发与校验。
// 注意：新 Vben 前端用 Bearer JWT（无状态），与旧 PHP 的 Redis serialize
// token 机制相互独立，双跑时共享的是 Mongo 数据而非会话。
package jwtauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是后台管理 token 的载荷。
type Claims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	RoleID   int64  `json:"role_id"`
	jwt.RegisteredClaims
}

// Manager 封装签发/解析。
type Manager struct {
	secret        []byte
	expireSeconds int64
}

func New(secret string, expireSeconds int64) *Manager {
	if expireSeconds <= 0 {
		expireSeconds = 36000
	}
	return &Manager{secret: []byte(secret), expireSeconds: expireSeconds}
}

// Generate 为一个管理员签发 token。
func (m *Manager) Generate(adminID int64, username string, roleID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		AdminID:  adminID,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expireSeconds) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse 校验并解析 token。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
