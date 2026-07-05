package admin

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"yhdm_service/internal/crud"
)

// RegisterSimpleModules 用通用 CRUD 引擎批量注册「纯增删改查」的简单模块。
// 复杂模块（视频/漫画/小说主体、支付、审核、媒资同步、统计、配置多分组等）
// 不在此列，单独实现。字段严格对齐 yhdm-php 各 Model 的 @property。
func RegisterSimpleModules(auth *gin.RouterGroup, db *mongo.Database) {
	// 增删改查
	reg := func(coll, listPath, name, unique string, search []SearchField, edit, required []string) {
		h := NewGenericHandler(crud.New[bson.M](db, coll), ModuleConfig{
			Collection: coll, Search: search, Edit: edit, Required: required, Unique: unique,
		})
		h.Register(auth, listPath, name)
	}
	// 只读列表（日志类）
	regList := func(coll, listPath string, search []SearchField) {
		h := NewGenericHandler(crud.New[bson.M](db, coll), ModuleConfig{Collection: coll, Search: search})
		auth.GET("/app/"+listPath, h.List)
	}
	name := func(p, f string, regex, isInt bool) SearchField {
		return SearchField{Param: p, Field: f, Regex: regex, Int: isInt}
	}

	// ---------- 视频管理（分类/标签/模块/每日/专题/关键字/玩法） ----------
	reg("movie_category", "movie-categories", "movie-category", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "position", "is_hot"}, []string{"name"})
	reg("movie_tag", "movie-tags", "movie-tag", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "parent_id", "attribute", "series", "is_hot", "count"}, []string{"name"})
	reg("movie_block", "movie-blocks", "movie-block", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "style", "is_disabled", "sort", "ico", "filter", "num", "position"}, []string{"name"})
	reg("movie_day", "movie-days", "movie-day", "",
		[]SearchField{name("label", "label", true, false)},
		[]string{"movie_id", "label"}, []string{"movie_id"})
	reg("movie_special", "movie-specials", "movie-special", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "img", "bg_img", "position", "description", "sort", "filter", "is_disabled"}, []string{"name"})
	reg("movie_keywords", "movie-keywords", "movie-keyword", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "is_hot", "sort", "num"}, []string{"name"})
	reg("play", "plays", "play", "",
		[]SearchField{name("title", "title", true, false), name("type", "type", false, false)},
		[]string{"title", "number", "tag", "city", "type", "description", "img_x", "video", "download_link"}, []string{"title"})

	// ---------- 漫画管理（标签/模块/关键字） ----------
	reg("comics_tag", "comics-tags", "comics-tag", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "attribute", "is_hot", "count"}, []string{"name"})
	reg("comics_block", "comics-blocks", "comics-block", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "style", "is_disabled", "sort", "ico", "filter", "num", "position"}, []string{"name"})
	reg("comics_keywords", "comics-keywords", "comics-keyword", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "is_hot", "sort", "num"}, []string{"name"})

	// ---------- 小说管理（标签/模块/关键字），注意 Novel_block 大写 N ----------
	reg("novel_tag", "novel-tags", "novel-tag", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "attribute", "is_hot", "count"}, []string{"name"})
	reg("Novel_block", "novel-blocks", "novel-block", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "style", "is_disabled", "sort", "ico", "filter", "num", "position"}, []string{"name"})
	reg("novel_keywords", "novel-keywords", "novel-keyword", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "is_hot", "sort", "num"}, []string{"name"})

	// ---------- 用户管理（用户组/任务） ----------
	reg("user_group", "user-groups", "user-group", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "description", "is_disabled", "sort", "level", "group", "promotion_type", "rate", "coupon_num", "price", "old_price"}, []string{"name"})
	reg("user_task", "user-tasks", "user-task", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "type", "description", "max_num"}, []string{"name"})

	// ---------- 营销管理（渠道包/充值商品/金刚区） ----------
	reg("channel_app", "channel-apps", "channel-app", "code",
		[]SearchField{name("name", "name", true, false), name("code", "code", true, false)},
		[]string{"name", "code", "link", "is_disabled"}, []string{"name", "code"})
	reg("product", "products", "product", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "type", "num", "gift_num", "vip_num", "price", "sort", "description", "price_tips", "is_disabled"}, []string{"name"})
	reg("kingkong", "kingkongs", "kingkong", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "position", "icon", "link", "block_position_code", "sort", "is_disabled"}, []string{"name"})

	// ---------- 系统设置（域名/快捷回复） ----------
	reg("domain", "domains", "domain", "",
		[]SearchField{name("url", "url", true, false), name("type", "type", false, false)},
		[]string{"url", "status", "type", "remark"}, []string{"url"})
	reg("quick_reply", "quick-replies", "quick-reply", "",
		[]SearchField{name("name", "name", true, false)},
		[]string{"name", "content", "sort"}, []string{"name"})

	// ---------- 日志记录（只读列表） ----------
	regList("admin_log", "admin-logs", []SearchField{name("admin_name", "admin_name", true, false)})
	regList("sms_log", "sms-logs", []SearchField{name("phone", "phone", true, false)})
}
