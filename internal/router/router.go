package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	adminHandler "yhdm_service/internal/handler/admin"
	"yhdm_service/internal/middleware"
	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// Deps 是路由所需的依赖集合。
type Deps struct {
	JWT       *jwtauth.Manager
	Auth      *service.AuthService
	AdminMgmt *service.AdminService
}

// New 构建 gin 引擎并注册全部路由。
// 前端 Vben 的 VITE_GLOB_API_URL = /api/backend，故业务路由挂在该前缀下。
func New(deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI: /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authH := adminHandler.NewAuthHandler(deps.Auth)

	api := r.Group("/api/backend")
	{
		// 公开接口
		api.POST("/app/login", authH.Login)
		api.POST("/logout", authH.Logout)

		// 需要登录的接口
		auth := api.Group("")
		auth.Use(middleware.Auth(deps.JWT))
		{
			auth.GET("/app/get-info", authH.GetInfo)
			auth.GET("/auth/codes", authH.Codes)
			auth.GET("/app/menus", authH.Menus)

			// 系统管理-管理员
			adminMgmtH := adminHandler.NewAdminMgmtHandler(deps.AdminMgmt)
			auth.GET("/app/admins", adminMgmtH.List)
			auth.GET("/app/options-admin-role", adminMgmtH.RoleOptions)
			auth.POST("/app/create-admin", adminMgmtH.Create)
			auth.POST("/app/update-admin", adminMgmtH.Update)
			auth.POST("/app/delete-admin", adminMgmtH.Delete)
			// token 刷新：当前 token 仍有效则重签一枚（简化实现，后续可换 refresh token）
			auth.POST("/auth/refresh", func(c *gin.Context) {
				newToken, err := deps.JWT.Generate(
					middleware.AdminID(c),
					c.GetString(middleware.CtxUsername),
					c.GetInt64(middleware.CtxRoleID),
				)
				if err != nil {
					response.Fail(c, "刷新失败")
					return
				}
				c.JSON(http.StatusOK, gin.H{"data": newToken, "status": 0})
			})
		}
	}

	return r
}
