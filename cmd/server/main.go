package main

import (
	"context"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"yhdm_service/internal/config"
	"yhdm_service/internal/crud"
	"yhdm_service/internal/database"
	_ "yhdm_service/internal/docs" // swagger 生成的文档（swag init 产出）
	"yhdm_service/internal/logger"
	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/repository"
	"yhdm_service/internal/router"
	"yhdm_service/internal/service"
)

// @title           yhdm_service API
// @version         1.0
// @description     樱花动漫后台服务 API（由 PHP/Phalcon 迁移至 Go/Gin）。
// @BasePath        /api/backend
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization

func main() {
	confPath := flag.String("conf", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*confPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	zlog, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer func() { _ = zlog.Sync() }()

	ctx := context.Background()

	mongoClient, db, err := database.NewMongo(ctx, cfg.Mongo)
	if err != nil {
		zlog.Fatal("连接 MongoDB 失败", zap.Error(err))
	}
	defer func() { _ = mongoClient.Disconnect(ctx) }()
	zlog.Info("MongoDB 已连接", zap.String("db", cfg.Mongo.Database))

	rdb, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		zlog.Fatal("连接 Redis 失败", zap.Error(err))
	}
	defer func() { _ = rdb.Close() }()
	zlog.Info("Redis 已连接", zap.String("addr", cfg.Redis.Addr))

	// 依赖装配（当前手工注入；模块变多后可引入 wire）。
	jwtMgr := jwtauth.New(cfg.JWT.Secret, cfg.JWT.ExpireSeconds)
	adminRepo := repository.NewAdminRepository(db)
	authSvc := service.NewAuthService(adminRepo, jwtMgr, cfg.App.Dev)
	adminMgmtSvc := service.NewAdminService(adminRepo)
	roleSvc := service.NewRoleService(adminRepo)
	authoritySvc := service.NewAuthorityService(adminRepo)

	// 文章管理：用通用 crud 基座
	articleSvc := service.NewArticleService(crud.New[model.Article](db, "article"))
	articleCategorySvc := service.NewArticleCategoryService(crud.New[model.ArticleCategory](db, "article_category"))
	blockPositionSvc := service.NewBlockPositionService(crud.New[model.BlockPosition](db, "block_position"))

	if !cfg.App.Dev {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := router.New(router.Deps{
		JWT: jwtMgr, Auth: authSvc, AdminMgmt: adminMgmtSvc, Role: roleSvc, Authority: authoritySvc,
		Article: articleSvc, ArticleCategory: articleCategorySvc, BlockPosition: blockPositionSvc,
		DB: db,
	})

	zlog.Info("yhdm_service 启动", zap.String("addr", cfg.App.Addr), zap.Bool("dev", cfg.App.Dev))
	if err := engine.Run(cfg.App.Addr); err != nil {
		zlog.Fatal("HTTP 服务退出", zap.Error(err))
	}
}
