# yhdm 后台迁移进度清单（验收用）

来源：yhdm-php `app/Controller/Backend/`（55 个控制器、约 238 个 action）。
每个模块按「实现 → 单元测试 → swagger → 真实数据冒烟 → 提交」推进。

图例：✅ 完成 ｜ 🚧 进行中 ｜ ⬜ 未开始 ｜ ⭐ 复杂(非纯CRUD，需专门设计)

---

## 全栈完成情况总览（后端 API + 前端页 + 菜单）
截至目前，以下菜单已**全栈可用**（登录 gotest/admin123 可见）：
- ✅ 系统管理：管理员 / 角色（复用 jh 页面）
- ✅ 文章管理：文章 / 分类 / 模块位置
- ✅ 用户管理：用户组 / 用户任务
- ✅ 视频管理：分类 / 标签 / 模块 / 专题 / 关键字
- ✅ 漫画管理：标签 / 模块 / 关键字
- ✅ 小说管理：标签 / 模块 / 关键字
- ✅ 营销管理：渠道包 / 充值商品 / 金刚区
- ✅ 系统设置：域名 / 快捷回复
- ✅ 日志记录：管理员日志 / 短信日志（只读）

**技术底座**：后端通用 CRUD 引擎(handler/admin/generic.go + modules.go)；前端通用组件
(components/CrudTable.vue + api/core/crud.ts)。简单模块两端均为配置驱动。

**待办（复杂模块，需单独设计，已跳过）**：
- 视频/漫画/小说**主体**(含章节、审核、媒资库同步、回收站) — MovieController/ComicsController/NovelController
- 用户主体 User(含后台充值+2FA+ES同步)、订单 Order、优惠券、兑换码、客服 Feedback
- 社区 Post/Comment(含 AI/审核)、活动 ActivityLand
- 渠道报表 Channel(统计+导出)、系统统计 System/Analysis、多分组配置 Config、任务调度 Job
- admin 的操作日志接口 admin-log、2FA generate/bind
- 视频每日推荐 movie-day、玩法 play（后端已注册，前端页未建）

---

## 已完成基础设施
- ✅ Gin 脚手架 / 配置 / 日志 / Mongo+Redis / 统一响应 / JWT / swagger / CORS
- ✅ 认证：login / get-info / auth-codes / menus / refresh / logout

## 系统用户（RBAC）
- ✅ **管理员 admin_user**：admins(列表) / options-admin-role / create / update / delete
- ✅ **角色 admin_role**：roles / permissions(权限树) / create / update / delete / save-permission
- ✅ **权限资源 authority**：authorities(列表) / authority-detail / save-authority / delete-authority
- ⬜ admin_user 的 do（启禁/改角色）— 旧 AdminUserController.do
- ⭐ 2FA：generate-google2fa / bind-google2fa（AdminUserController 相关）

## 系统统计 / 首页
- ⭐ SystemController：home / hour / dau / theme / errorLogs / password / clearCache / fake / uploadCsv / behavior（统计聚合+上传+缓存）
- ⭐ AnalysisController：movie / cartoon（统计聚合）
- ⬜ AccountController：list / credit（account_log 只读）

## 用户管理
- ⭐ UserController：list / detail / save / do / recharge / find / doFind（含后台充值+2FA+ES同步）
- ⬜ UserGroupController：list / detail / save（user_group）
- ⬜ UserTaskController：list / detail / save（user_task）
- ⬜ UserUpController：list / detail / save（user_up）
- ⬜ UserCouponController：list / detail / save（user_coupon）
- ⭐ UserCodeController：list / detail / save / export / log（批量生成+导出）
- ⭐ OrderController：vip / point / collection / buy（多集合订单聚合）
- ⭐ FeedbackController：list / vip / detail / message / save（客服会话）

## 视频管理
- ⭐ MovieController：list / warehouse / review / recycle / detail / moreLink / save / async / asyncCommon / do / update / updateLinks / widget（审核+媒资同步+CDN）
- ⬜ MovieCategoryController：list / detail / save（movie_category）
- ⬜ MovieTagController：list / detail / save / update（movie_tag）
- ⬜ MovieBlockController：list / detail / save（movie_block）
- ⬜ MovieDayController：list / detail / save / update（movie_day）
- ⬜ MovieSpecialController：list / detail / save（movie_special）
- ⬜ MovieKeywordsController：list / detail / save（movie_keywords）
- ⬜ PlayController：game / luoliao / yuepao / detail / save / update（play）

## 漫画管理
- ⭐ ComicsController：list / warehouse / recycle / detail / chapterDetail / save / do / update / async（章节+媒资同步）
- ⬜ ComicsTagController：list / detail / save / update
- ⬜ ComicsBlockController：list / detail / save
- ⬜ ComicsKeywordsController：list / detail / save

## 小说管理
- ⭐ NovelController：list / warehouse / recycle / detail / chapterDetail / save / do / update / async
- ⬜ NovelTagController：list / detail / save / update
- ⬜ NovelBlockController：list / detail / save（Novel_block）
- ⬜ NovelKeywordsController：list / detail / save

## 社区管理
- ⭐ PostController：list / ai / detail / save / update / changeFromMovie / asyncFromMrs（AI+转换+同步）
- ⬜ PostCategoryController：list / detail / save
- ⬜ PostBlockController：list / detail / save
- ⭐ CommentController：comics / movie / post / novel / comment / doComment / doReply / do（多业务评论审核）
- ⭐ CommentReplyController：reply / movie / post / cartoon / do

## 营销管理
- ⭐ ChannelController：list / detail / save / report / reportOne / exportReport / do / reportFree（渠道统计+导出）
- ⬜ ChannelAppController：list / detail / save / do（channel_app 渠道包）
- ⬜ ProductController：list / detail / save（product 充值商品）
- ⬜ KingkongController：list / detail / save / do（kingkong 金刚区）

## 活动专区
- ⬜ ActivityLandController：index / list / detail / save / do（activity_land）

## 文章管理
- ✅ ArticleController：articles(列表) / article-detail / save-article / delete-article
- ✅ ArticleCategoryController：article-categories / save-article-category / delete-article-category
- ✅ BlockPosController：block-positions / save-block-position / delete-block-position

## 系统设置
- ⭐ ConfigController：base / other / movie / apk / app / cdn / userTask / center / save（多分组配置）
- ⬜ DomainController：list / detail / save / do（domain）
- ⬜ AutoLineController：list（domain 只读）
- ⬜ QuickReplyController：list / detail / save（quick_reply）
- ⭐ JobController：list / do（任务调度）

## 日志记录
- ⬜ AdminLogsController：list / del（admin_log）
- ⬜ SmsLogsController：list（sms_log）

---

### 迁移策略备注
- 纯 CRUD（⬜ 大多数）：list/detail/save/delete，直接按 Model @property 字段迁移。
- 复杂项（⭐）：媒资同步(第三方)、支付充值、统计聚合、审核、文件上传导出、多分组配置、任务调度——后置或单独设计。
- 数据源：第一阶段共用旧库 Mongo（本机为 `manhua` 库）。字段语义见 yhdm-php 各 Model 的 @property 中文注释。
