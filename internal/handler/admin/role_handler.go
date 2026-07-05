package admin

import (
	"github.com/gin-gonic/gin"

	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// RoleHandler 处理"系统管理-角色"的增删改查与权限分配。
type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List 角色列表
// @Summary      角色列表
// @Tags         系统管理-角色
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=object}  "data: {list:[RoleListItem]}"
// @Router       /app/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Permissions 权限树
// @Summary      权限树(用于给角色分配权限)
// @Tags         系统管理-角色
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=object}  "data: {list:[PermissionNode]}"
// @Router       /app/permissions [get]
func (h *RoleHandler) Permissions(c *gin.Context) {
	list, err := h.svc.Permissions(c.Request.Context())
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

type createRoleReq struct {
	Name string `json:"name"`
}

// Create 新增角色
// @Summary      新增角色
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createRoleReq  true  "角色名"
// @Success      200   {object}  response.Body{data=object}  "data: {id, message}"
// @Router       /app/create-role [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req createRoleReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "message": "创建成功"})
}

type updateRoleReq struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Update 更新角色名
// @Summary      更新角色名
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      updateRoleReq  true  "角色ID与名称"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/update-role [post]
func (h *RoleHandler) Update(c *gin.Context) {
	var req updateRoleReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.svc.UpdateName(c.Request.Context(), req.ID, req.Name); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true, "message": "更新成功"})
}

type deleteRoleReq struct {
	ID int64 `json:"id"`
}

// Delete 删除角色
// @Summary      删除角色
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      deleteRoleReq  true  "角色ID"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/delete-role [post]
func (h *RoleHandler) Delete(c *gin.Context) {
	var req deleteRoleReq
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

type savePermissionReq struct {
	ID             int64  `json:"id"`
	SiteID         int64  `json:"site_id"`
	PermissionList string `json:"permission_list"`
}

// SavePermission 给角色分配权限
// @Summary      保存角色权限
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      savePermissionReq  true  "角色ID与权限列表(逗号分隔)"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/save-permission [post]
func (h *RoleHandler) SavePermission(c *gin.Context) {
	var req savePermissionReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.svc.SavePermission(c.Request.Context(), req.ID, req.PermissionList); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true, "message": "保存成功"})
}
