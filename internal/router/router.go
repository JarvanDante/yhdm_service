package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"

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
	Role      *service.RoleService
	Authority *service.AuthorityService

	// 文章管理
	Article         *service.ArticleService
	ArticleCategory *service.ArticleCategoryService
	BlockPosition   *service.BlockPositionService

	// 通用简单模块用（视频/漫画/小说/用户/营销/系统设置/日志的简单 CRUD）
	DB *mongo.Database
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

			// 系统管理-角色
			roleH := adminHandler.NewRoleHandler(deps.Role)
			auth.GET("/app/roles", roleH.List)
			auth.GET("/app/permissions", roleH.Permissions)
			auth.POST("/app/create-role", roleH.Create)
			auth.POST("/app/update-role", roleH.Update)
			auth.POST("/app/delete-role", roleH.Delete)
			auth.POST("/app/save-permission", roleH.SavePermission)

			// 系统管理-权限资源(菜单节点)
			authorityH := adminHandler.NewAuthorityHandler(deps.Authority)
			auth.GET("/app/authorities", authorityH.List)
			auth.GET("/app/authority-detail", authorityH.Detail)
			auth.POST("/app/save-authority", authorityH.Save)
			auth.POST("/app/delete-authority", authorityH.Delete)

			// 文章管理（文章 / 分类 / 模块位置）
			articleH := adminHandler.NewArticleHandler(deps.Article, deps.ArticleCategory, deps.BlockPosition)
			auth.GET("/app/articles", articleH.ArticleList)
			auth.GET("/app/article-detail", articleH.ArticleDetail)
			auth.POST("/app/save-article", articleH.SaveArticle)
			auth.POST("/app/delete-article", articleH.DeleteArticle)
			auth.GET("/app/article-categories", articleH.CategoryList)
			auth.POST("/app/save-article-category", articleH.SaveCategory)
			auth.POST("/app/delete-article-category", articleH.DeleteCategory)
			auth.GET("/app/block-positions", articleH.BlockList)
			auth.POST("/app/save-block-position", articleH.SaveBlock)
			auth.POST("/app/delete-block-position", articleH.DeleteBlock)

			// 通用简单模块（视频/漫画/小说/用户/营销/系统设置/日志）
			if deps.DB != nil {
				adminHandler.RegisterSimpleModules(auth, deps.DB)
			}
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
