package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

type TicketHandler struct {
	tickets *service.TicketService
	catalog *service.CatalogService
}

func NewTicketHandler(tickets *service.TicketService, catalog *service.CatalogService) *TicketHandler {
	return &TicketHandler{tickets: tickets, catalog: catalog}
}

func (h *TicketHandler) Create(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	var req model.CreateTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Create(uid, role, req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) List(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	page, pageSize := parsePage(c)
	q := model.TicketListQuery{
		Page:      page,
		PageSize:  pageSize,
		Status:    c.Query("status"),
		Category:  c.Query("category"),
		Priority:  c.Query("priority"),
		Q:         c.Query("q"),
		Scope:     c.Query("scope"),
		TimeField: c.Query("time_field"),
		FromDay:   c.Query("from"),
		ToDay:     c.Query("to"),
	}
	if raw := c.Query("assignee_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "assignee_id 不合法")
			return
		}
		q.AssigneeID = &id
	}
	out, err := h.tickets.List(uid, role, q)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) Stats(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	out, err := h.tickets.Stats(uid, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) GetDetail(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	ticket, err := h.tickets.GetDetail(uid, id, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, ticket)
}

func (h *TicketHandler) Assign(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req model.AssignTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Assign(uid, id, role, req.AssigneeID)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) Claim(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	t, err := h.tickets.Claim(uid, id, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) Update(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req model.TicketActionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Update(uid, id, role, req.Action, req.Reason)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) Comment(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req model.CreateCommentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.tickets.Comment(uid, id, role, req.Body)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) Transfer(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req model.TransferTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Transfer(uid, id, role, req.AssigneeID, req.Reason)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) Board(c *gin.Context) {
	_, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	out, err := h.tickets.Board(role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) CloseStale(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	out, err := h.tickets.CloseStaleResolved(uid, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) Edit(c *gin.Context) {
	uid, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req model.EditTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Edit(uid, id, role, req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TicketHandler) ListAgents(c *gin.Context) {
	_, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	if role != "agent" && role != "admin" {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "没有权限查看 IT 列表")
		return
	}
	out, err := h.tickets.ListAgents()
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) ListTemplates(c *gin.Context) {
	if _, _, ok := actor(c); !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	out, err := h.catalog.ListTemplates(true)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *TicketHandler) ListCannedReplies(c *gin.Context) {
	_, role, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	if role != "agent" && role != "admin" {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "只有 IT 和管理员能用常用回复")
		return
	}
	out, err := h.catalog.ListReplies(true)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "工单 id 不合法")
		return 0, false
	}
	return id, true
}
