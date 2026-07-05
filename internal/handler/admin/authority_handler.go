package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"yhdm_service/internal/repository"
	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// AuthorityHandler 处理"系统管理-权限资源"的增删改查。
type AuthorityHandler struct {
	svc *service.AuthorityService
}

func NewAuthorityHandler(svc *service.AuthorityService) *AuthorityHandler {
	return &AuthorityHandler{svc: svc}
}

// List 权限资源列表
// @Summary      权限资源(菜单节点)列表(分页)
// @Tags         系统管理-权限资源
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int     false  "页码"  default(1)
// @Param        size       query     int     false  "每页"  default(15)
// @Param        name       query     string  false  "名称模糊"
// @Param        key        query     string  false  "标识模糊"
// @Param        parent_id  query     int     false  "上级ID"
// @Param        is_menu    query     int     false  "是否菜单:1是/0否"
// @Success      200        {object}  response.Body{data=service.AuthorityListResult}
// @Router       /app/authorities [get]
func (h *AuthorityHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "15"))

	f := repository.AuthorityFilter{
		Name: c.Query("name"),
		Key:  c.Query("key"),
	}
	if v := c.Query("parent_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.ParentID = &n
		}
	}
	if v := c.Query("is_menu"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.IsMenu = &n
		}
	}

	res, err := h.svc.List(c.Request.Context(), f, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, res)
}

// Detail 权限资源详情
// @Summary      权限资源详情
// @Tags         系统管理-权限资源
// @Produce      json
// @Security     BearerAuth
// @Param        id   query     int  true  "权限资源ID"
// @Success      200  {object}  response.Body{data=service.AuthorityListItem}
// @Router       /app/authority-detail [get]
func (h *AuthorityHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	a, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, a)
}

// saveAuthorityReq 新增/更新权限资源（id>0 为更新）。
type saveAuthorityReq struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	ParentID  int64  `json:"parent_id"`
	Sort      int    `json:"sort"`
	ClassName string `json:"class_name"`
	IsMenu    int    `json:"is_menu"`
	Link      string `json:"link"`
}

// Save 新增或更新权限资源
// @Summary      新增/更新权限资源(id>0 为更新)
// @Tags         系统管理-权限资源
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      saveAuthorityReq  true  "权限资源信息"
// @Success      200   {object}  response.Body{data=object}  "data: {id, message}"
// @Router       /app/save-authority [post]
func (h *AuthorityHandler) Save(c *gin.Context) {
	var req saveAuthorityReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.svc.Save(c.Request.Context(), service.AuthorityInput{
		ID: req.ID, Name: req.Name, Key: req.Key, ParentID: req.ParentID,
		Sort: req.Sort, ClassName: req.ClassName, IsMenu: req.IsMenu, Link: req.Link,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "message": "保存成功"})
}

type deleteAuthorityReq struct {
	ID int64 `json:"id"`
}

// Delete 删除权限资源
// @Summary      删除权限资源
// @Tags         系统管理-权限资源
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      deleteAuthorityReq  true  "权限资源ID"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/delete-authority [post]
func (h *AuthorityHandler) Delete(c *gin.Context) {
	var req deleteAuthorityReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), req.ID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true, "message": "删除成功"})
}
