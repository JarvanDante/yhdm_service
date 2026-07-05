package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 是与前端 Vben(web-ele) 约定的统一响应结构。
// 关键：成功码为 0（前端 defaultResponseInterceptor 里 successCode: 0）。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	CodeSuccess = 0
	CodeError   = -1
)

// OK 返回成功响应，HTTP 状态恒为 200，业务码为 0。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: CodeSuccess, Message: "success", Data: data})
}

// Fail 返回业务失败（HTTP 200 + code!=0），前端据 code 弹错误提示。
func Fail(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Body{Code: CodeError, Message: message, Data: nil})
}

// FailCode 允许自定义业务码（如 401 交给前端触发重登）。
func FailCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Body{Code: code, Message: message, Data: nil})
}

// Unauthorized 用 HTTP 401，触发前端 authenticateResponseInterceptor 的重登逻辑。
func Unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, Body{Code: http.StatusUnauthorized, Message: message, Data: nil})
}
