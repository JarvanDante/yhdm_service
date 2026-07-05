package admin

import (
	"github.com/gin-gonic/gin"

	"yhdm_service/internal/middleware"
	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// AuthHandler 处理后台登录/用户信息/权限/菜单等接口。
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// LoginReq 兼容 JSON 与 form 两种提交（Vben 默认发 JSON）。
type LoginReq struct {
	Username   string `json:"username" form:"username" example:"admin"`
	Password   string `json:"password" form:"password" example:"admin123"`
	GoogleCode string `json:"google_code" form:"google_code" example:"123456"`
	// Vben 登录组件里的验证码字段名可能是 code，这里一并接住。
	Code string `json:"code" form:"code"`
}

// Login 后台登录
// @Summary      后台管理员登录
// @Description  用户名+密码(+谷歌验证码)登录，成功返回 JWT。dev 模式跳过 2FA。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      LoginReq  true  "登录参数"
// @Success      200   {object}  response.Body{data=object}  "data: {token, socket}"
// @Router       /app/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.Username == "" || req.Password == "" {
		response.Fail(c, "参数错误")
		return
	}
	googleCode := req.GoogleCode
	if googleCode == "" {
		googleCode = req.Code
	}

	token, _, err := h.auth.Login(c.Request.Context(), req.Username, req.Password, googleCode, c.ClientIP())
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	// socket 字段为旧前端 IM 长连接地址，本期不接 IM，返回空串占位。
	response.OK(c, gin.H{"token": token, "socket": ""})
}

// GetInfo 当前管理员信息
// @Summary      获取当前登录管理员信息
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=service.UserInfo}
// @Router       /app/get-info [get]
func (h *AuthHandler) GetInfo(c *gin.Context) {
	info, err := h.auth.GetUserInfo(c.Request.Context(), middleware.AdminID(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, info)
}

// Codes 权限码数组
// @Summary      获取当前管理员的权限码(按钮级)
// @Description  超级管理员返回 ["*"]
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=[]string}
// @Router       /auth/codes [get]
func (h *AuthHandler) Codes(c *gin.Context) {
	codes, err := h.auth.AccessCodes(c.Request.Context(), middleware.AdminID(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, codes)
}

// Menus 菜单树
// @Summary      获取当前管理员可见的菜单树
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=[]service.MenuItem}
// @Router       /app/menus [get]
func (h *AuthHandler) Menus(c *gin.Context) {
	menus, err := h.auth.Menus(c.Request.Context(), middleware.AdminID(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, menus)
}

// Logout 退出登录
// @Summary      退出登录
// @Tags         认证
// @Produce      json
// @Success      200  {object}  response.Body
// @Router       /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 无状态，退出由前端清除 token 即可；此处返回成功占位。
	response.OK(c, gin.H{"success": true})
}
