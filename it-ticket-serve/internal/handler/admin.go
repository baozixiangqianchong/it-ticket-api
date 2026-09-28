package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/middleware"
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

// AdminHandler 管理员接口。路由已挂 JWT + RequireAdmin，这里不再判角色。
type AdminHandler struct {
	admin   *service.AdminService
	catalog *service.CatalogService
}

func NewAdminHandler(admin *service.AdminService, catalog *service.CatalogService) *AdminHandler {
	return &AdminHandler{admin: admin, catalog: catalog}
}

// ListUsers GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	out, err := h.admin.ListUsers(page, pageSize, c.Query("role"), c.Query("status"), c.Query("q"))
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
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	u, err := h.admin.UpdateRole(actorID, id, req.Role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "用户 id 不合法")
		return
	}
	var req model.UpdateUserStatusInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	u, err := h.admin.UpdateStatus(actorID, id, req.Status)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

func (h *AdminHandler) CreateInvite(c *gin.Context) {
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	out, err := h.admin.CreateInvite(actorID)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) ListInvites(c *gin.Context) {
	out, err := h.admin.ListInvites()
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) ListAudits(c *gin.Context) {
	page, pageSize := parsePage(c)
	out, err := h.admin.ListActivity(page, pageSize, c.Query("kind"), c.Query("action"), c.Query("q"), c.Query("from"), c.Query("to"))
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) ListTemplates(c *gin.Context) {
	out, err := h.catalog.ListTemplates(false)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) CreateTemplate(c *gin.Context) {
	var req model.TemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.catalog.CreateTemplate(req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) UpdateTemplate(c *gin.Context) {
	id, ok := parseID(c, "模板 id 不合法")
	if !ok {
		return
	}
	var req model.TemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.catalog.UpdateTemplate(id, req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) DeleteTemplate(c *gin.Context) {
	id, ok := parseID(c, "模板 id 不合法")
	if !ok {
		return
	}
	if err := h.catalog.DeleteTemplate(id); err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *AdminHandler) ListReplies(c *gin.Context) {
	out, err := h.catalog.ListReplies(false)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) CreateReply(c *gin.Context) {
	var req model.CannedReplyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.catalog.CreateReply(req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) UpdateReply(c *gin.Context) {
	id, ok := parseID(c, "常用回复 id 不合法")
	if !ok {
		return
	}
	var req model.CannedReplyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.catalog.UpdateReply(id, req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AdminHandler) DeleteReply(c *gin.Context) {
	id, ok := parseID(c, "常用回复 id 不合法")
	if !ok {
		return
	}
	if err := h.catalog.DeleteReply(id); err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func parseID(c *gin.Context, message string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, message)
		return 0, false
	}
	return id, true
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
