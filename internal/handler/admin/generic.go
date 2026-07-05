package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/crud"
	"yhdm_service/internal/response"
)

// 通用 CRUD 引擎：配置驱动，适用于「纯增删改查」的简单集合，避免每个模块
// 重复写 model/service/handler。复杂模块（支付、审核、媒资同步、统计等）仍单独实现。

// SearchField 描述一个列表筛选字段。
type SearchField struct {
	Param string // 前端查询参数名
	Field string // mongo 字段名
	Regex bool   // true=模糊匹配，false=精确
	Int   bool   // 值按整数解析
}

// ModuleConfig 描述一个简单 CRUD 模块。
type ModuleConfig struct {
	Collection  string        // mongo 集合名
	Search      []SearchField // 列表筛选字段
	Edit        []string      // 可保存字段白名单（前端 json key 与 mongo 字段同名）
	Required    []string      // 保存时必填字段
	Unique      string        // 唯一字段（如 code），为空则不校验
	DefaultSort bson.D        // 默认排序，空则按 _id 降序
	DefaultSize int           // 默认每页
}

// GenericHandler 是通用 CRUD 处理器，依赖 crud.Repo 便于单测注入 MemStore。
type GenericHandler struct {
	repo crud.Repo[bson.M]
	cfg  ModuleConfig
}

// NewGenericHandler 用给定仓储与配置构造处理器。
func NewGenericHandler(repo crud.Repo[bson.M], cfg ModuleConfig) *GenericHandler {
	if cfg.DefaultSize <= 0 {
		cfg.DefaultSize = 15
	}
	return &GenericHandler{repo: repo, cfg: cfg}
}

// List 分页列表。
func (h *GenericHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", strconv.Itoa(h.cfg.DefaultSize)))
	query := bson.M{}
	for _, sf := range h.cfg.Search {
		v := c.Query(sf.Param)
		if v == "" {
			continue
		}
		switch {
		case sf.Regex:
			query[sf.Field] = bson.M{"$regex": v, "$options": "i"}
		case sf.Int:
			if n, err := strconv.Atoi(v); err == nil {
				query[sf.Field] = n
			}
		default:
			query[sf.Field] = v
		}
	}
	list, total, err := h.repo.List(c.Request.Context(), query, h.cfg.DefaultSort, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	if list == nil {
		list = []bson.M{}
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// Save 新增/更新（body.id>0 为更新）。
func (h *GenericHandler) Save(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	for _, r := range h.cfg.Required {
		if isEmptyVal(body[r]) {
			response.Fail(c, r+" 不能为空")
			return
		}
	}
	ctx := c.Request.Context()
	id := toInt64(body["id"])

	// 唯一性校验（排除自身）
	if h.cfg.Unique != "" {
		if uv, ok := body[h.cfg.Unique]; ok && !isEmptyVal(uv) {
			dup, err := h.repo.FindOne(ctx, bson.M{h.cfg.Unique: uv})
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			if dup != nil && toInt64((*dup)["_id"]) != id {
				response.Fail(c, "该"+h.cfg.Unique+"已存在")
				return
			}
		}
	}

	// 组装可写字段
	set := bson.M{}
	for _, f := range h.cfg.Edit {
		if v, ok := body[f]; ok {
			set[f] = v
		}
	}
	now := time.Now().Unix()
	set["updated_at"] = now

	if id > 0 {
		cur, err := h.repo.FindByID(ctx, id)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		if cur == nil {
			response.Fail(c, "数据不存在")
			return
		}
		if err := h.repo.Update(ctx, id, set); err != nil {
			response.Fail(c, err.Error())
			return
		}
		response.OK(c, gin.H{"id": id, "message": "保存成功"})
		return
	}

	newID, err := crud.Create[bson.M](ctx, h.repo, func(nid int64) any {
		doc := bson.M{"_id": nid, "created_at": now}
		for k, v := range set {
			doc[k] = v
		}
		return doc
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": newID, "message": "保存成功"})
}

// Delete 按 id 删除。
func (h *GenericHandler) Delete(c *gin.Context) {
	var req idReq
	if err := c.ShouldBind(&req); err != nil || req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	n, err := h.repo.Delete(c.Request.Context(), req.ID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	if n == 0 {
		response.Fail(c, "数据不存在")
		return
	}
	response.OK(c, gin.H{"success": true, "message": "删除成功"})
}

// Register 把 list/save/delete 注册到给定路由组，路径形如
// GET /app/<listPath>、POST /app/save-<name>、POST /app/delete-<name>。
func (h *GenericHandler) Register(g *gin.RouterGroup, listPath, name string) {
	g.GET("/app/"+listPath, h.List)
	g.POST("/app/save-"+name, h.Save)
	g.POST("/app/delete-"+name, h.Delete)
}

func isEmptyVal(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	return false
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}
