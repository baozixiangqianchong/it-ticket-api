package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

// AdminHandler 管理员接口。路由已挂 JWT + RequireAdmin，这里不再判角色。
type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// ListUsers GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	out, err := h.admin.ListUsers(page, pageSize)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateRole POST /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "用户 id 不合法")
		return
	}
	var req model.UpdateRoleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	u, err := h.admin.UpdateRole(id, req.Role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

// parsePage 读 query。page 从 1 开始，默认 1；page_size 默认 20，最大 50。
func parsePage(c *gin.Context) (page, pageSize int) {
	page, pageSize = 1, 20
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.Query("page_size")); err == nil && v > 0 {
		pageSize = v
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}
