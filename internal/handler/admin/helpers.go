package admin

import (
	"context"

	"github.com/gin-gonic/gin"

	"yhdm_service/internal/response"
)

// idReq 是通用的 {id} 请求体（删除、按 id 操作等）。
type idReq struct {
	ID int64 `json:"id"`
}

// respond 统一处理「查询类」接口的返回。
func respond(c *gin.Context, data any, err error) {
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, data)
}

// respondSave 统一处理「保存类」接口的返回（返回新/旧 id）。
func respondSave(c *gin.Context, id int64, err error) {
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "message": "保存成功"})
}

// deleteByID 统一处理「按 id 删除」接口。
func deleteByID(c *gin.Context, del func(context.Context, int64) error) {
	var req idReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := del(c.Request.Context(), req.ID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true, "message": "删除成功"})
}
