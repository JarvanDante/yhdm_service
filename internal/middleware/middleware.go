package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/response"
)

// gin.Context 中存放当前管理员信息的键。
const (
	CtxAdminID  = "admin_id"
	CtxUsername = "username"
	CtxRoleID   = "role_id"
)

// CORS 允许跨域（开发期前端 vite 独立端口直连）。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,Accept-Language")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// Auth 校验 Bearer JWT，并把管理员信息注入 context。
func Auth(jwtMgr *jwtauth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Unauthorized(c, "未登录")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer"))
		claims, err := jwtMgr.Parse(token)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			return
		}
		c.Set(CtxAdminID, claims.AdminID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxRoleID, claims.RoleID)
		c.Next()
	}
}

// AdminID 从 context 取当前管理员 ID。
func AdminID(c *gin.Context) int64 {
	if v, ok := c.Get(CtxAdminID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
