package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// AdminMgmtHandler 处理"系统管理-管理员"的增删改查。
type AdminMgmtHandler struct {
	svc *service.AdminService
}

func NewAdminMgmtHandler(svc *service.AdminService) *AdminMgmtHandler {
	return &AdminMgmtHandler{svc: svc}
}

// List 管理员列表
// @Summary      管理员列表(分页)
// @Tags         系统管理-管理员
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "页码"      default(1)
// @Param        size      query     int     false  "每页数量"  default(10)
// @Param        username  query     string  false  "用户名模糊筛选"
// @Param        status    query     string  false  "状态筛选:1启用/0禁用"
// @Success      200       {object}  response.Body{data=service.AdminListResult}
// @Router       /app/admins [get]
func (h *AdminMgmtHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	username := c.Query("username")

	var status *int
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = &v
		}
	}

	res, err := h.svc.List(c.Request.Context(), username, status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, res)
}

// RoleOptions 角色下拉选项
// @Summary      角色下拉选项
// @Tags         系统管理-管理员
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=[]service.RoleOption}
// @Router       /app/options-admin-role [get]
func (h *AdminMgmtHandler) RoleOptions(c *gin.Context) {
	opts, err := h.svc.RoleOptions(c.Request.Context())
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, opts)
}

// createAdminReq 对齐前端 CreateAdminParams。
type createAdminReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Role     int64  `json:"role"`
	Status   int    `json:"status"`
}

// Create 新增管理员
// @Summary      新增管理员
// @Tags         系统管理-管理员
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createAdminReq  true  "管理员信息"
// @Success      200   {object}  response.Body{data=object}  "data: {id, message}"
// @Router       /app/create-admin [post]
func (h *AdminMgmtHandler) Create(c *gin.Context) {
	var req createAdminReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.svc.Create(c.Request.Context(), service.CreateAdminInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "message": "创建成功"})
}

// updateAdminReq 对齐前端 UpdateAdminParams。
type updateAdminReq struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	RoleID   int64  `json:"role_id"`
	Status   int    `json:"status"`
	Password string `json:"password"`
}

// Update 更新管理员
// @Summary      更新管理员
// @Tags         系统管理-管理员
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      updateAdminReq  true  "管理员信息(password 为空则不改密码)"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/update-admin [post]
func (h *AdminMgmtHandler) Update(c *gin.Context) {
	var req updateAdminReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if req.ID <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	err := h.svc.Update(c.Request.Context(), service.UpdateAdminInput{
		ID:       req.ID,
		Username: req.Username,
		Nickname: req.Nickname,
		RoleID:   req.RoleID,
		Status:   req.Status,
		Password: req.Password,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true, "message": "更新成功"})
}

// deleteAdminReq 对齐前端 DeleteAdminParams。
type deleteAdminReq struct {
	ID int64 `json:"id"`
}

// Delete 删除管理员
// @Summary      删除管理员
// @Tags         系统管理-管理员
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      deleteAdminReq  true  "管理员ID"
// @Success      200   {object}  response.Body{data=object}  "data: {success, message}"
// @Router       /app/delete-admin [post]
func (h *AdminMgmtHandler) Delete(c *gin.Context) {
	var req deleteAdminReq
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
