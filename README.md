# yhdm_service

樱花动漫（yhdm）后台服务，由原 PHP（Phalcon 4）迁移而来。独立 **Gin HTTP** 服务，前端为 Vben Admin（`yhdm_backend`）直连。

## 迁移策略

- **第一步（当前）**：Go 直接连旧系统的 **MongoDB**，与旧 PHP 共用同一份数据，支持新旧双跑灰度、逐模块验证正确性。
- **第二步（后续）**：Go 跑稳后，再单独做 MongoDB → MySQL 的数据与代码迁移。
- IM（自研 IM SDK）本期**不接**。

## 技术栈

Gin · mongo-go-driver · go-redis/v9 · golang-jwt/v5 · pquerna/otp(2FA) · viper · zap

## 目录结构

```
cmd/server           程序入口
configs              config.yaml(生产) / config.dev.yaml(开发)
internal/
  config             配置加载(viper，支持 YHDM_ 前缀环境变量覆盖)
  logger             zap 日志
  database           mongo / redis 连接
  model              Mongo 集合模型(字段对齐旧 PHP)
  repository         数据访问层
  service            业务逻辑层
  handler/admin      HTTP 处理器
  middleware         JWT 鉴权 / CORS
  response           统一响应 {code,message,data}(成功码=0，对齐 Vben)
  router             路由注册(/api/backend 前缀)
  pkg/
    password         复刻旧密码哈希 md5(md5(pwd)+"This is password"+md5(slat))
    jwtauth          JWT 签发/校验
    totp             谷歌 2FA 校验
```

## 已实现（认证竖切）

对齐 Vben(web-ele) 前端契约：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/backend/app/login` | 登录，返回 `{token, socket}` |
| GET  | `/api/backend/app/get-info` | 当前管理员信息 |
| GET  | `/api/backend/auth/codes` | 权限码数组（超管为 `["*"]`） |
| GET  | `/api/backend/app/menus` | 菜单树（对齐 authority 两级结构） |
| POST | `/api/backend/auth/refresh` | 刷新 token |
| POST | `/api/backend/logout` | 退出 |
| GET  | `/health` | 健康检查 |

数据来源：Mongo 集合 `admin_user` / `admin_role` / `authority`。密码与现有账号完全兼容。

## 运行

```bash
make tidy      # 拉依赖
make dev       # 用 configs/config.dev.yaml 启动(开发模式跳过 2FA)
```

先按实际环境改 `configs/config.dev.yaml` 里的 Mongo/Redis 地址与库名。

## 待办（下一步）

- [ ] 登录 IP 白名单校验（旧系统读 `configs` 集合的 `whitelist_ip`）
- [ ] admin_user / admin_role / authority 的增删改查（后台管理页）
- [ ] 内容业务模块：动漫/漫画/小说/文章、用户、支付、活动……
- [ ] 定时任务(asynq/cron)、ES 搜索、Excel 导出
