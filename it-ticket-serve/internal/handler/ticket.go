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

// TicketHandler 处理工单相关 HTTP。
// 这里只负责：读参数、取当前用户、调 service、写响应。不开事务、不拼 SQL。
type TicketHandler struct {
	tickets *service.TicketService
}

func NewTicketHandler(tickets *service.TicketService) *TicketHandler {
	return &TicketHandler{tickets: tickets}
}

// Create POST /api/v1/tickets
// 路由已挂 JWT。中间件验过 token 后把 uid 放进 context，这里取出来当创建人。
func (h *TicketHandler) Create(c *gin.Context) {
	// UID 是中间件写入的，不是请求体里的字段。客户端不能冒充别人提单。
	uid, ok := middleware.UID(c)
	if !ok {
		// 正常情况走不到：没过 JWT 中间件根本进不了这个 handler。
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}

	var req model.CreateTicketInput
	// 只绑 title / description / category。status、creator_id 就算客户端传了也会被忽略。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}

	t, err := h.tickets.Create(uid, req)
	if err != nil {
		writeAPIError(c, err) // 400 校验失败走 service.Error；其它当 500，不把 SQL 回给客户端
		return
	}
	response.OK(c, t) // data 是刚建的工单：status=open，assignee_id=null
}

// List GET /api/v1/tickets
// 路由已挂 JWT。中间件验过 token 后把 uid 放进 context，这里取出来当创建人。
func (h *TicketHandler) List(c *gin.Context) {
	uid, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	// 获取当前用户角色
	role, ok := middleware.Role(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}

	tickets, err := h.tickets.List(uid, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, tickets)
}

// GetDetail GET /api/v1/tickets/:id
// 和列表一样带上 uid、role；id 来自路径，必须是正整数。
func (h *TicketHandler) GetDetail(c *gin.Context) {
	uid, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	role, ok := middleware.Role(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "工单 id 不合法")
		return
	}
	ticket, err := h.tickets.GetDetail(uid, id, role)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, ticket)
}

// Assign POST /api/v1/tickets/:id/assign
// 路径上的 id 是工单；JSON 的 assignee_id 是受理人（IT）。
// 管理员 uid 只当操作人写审计，不会写进 assignee_id。
func (h *TicketHandler) Assign(c *gin.Context) {
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	role, ok := middleware.Role(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	// 获取路径上的 id，工单id
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "工单 id 不合法")
		return
	}
	// 获取请求体中的 assignee_id，受理人id
	var req model.AssignTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Assign(actorID, ticketID, role, req.AssigneeID)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

// Update POST /api/v1/tickets/:id/update
// 用 action 改状态，不接收客户端直接写的 status。
func (h *TicketHandler) Update(c *gin.Context) {
	// 获取当前用户id
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	role, ok := middleware.Role(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	// 获取路径上的 id，工单id
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "工单 id 不合法")
		return
	}
	// 获取请求体中的 action，动作
	var req model.TicketActionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	t, err := h.tickets.Update(actorID, ticketID, role, req.Action)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, t)
}

// Comment POST /api/v1/tickets/:id/comments
// 路径上的 id 是工单；JSON 的 body 是留言。作者是当前登录用户。
func (h *TicketHandler) Comment(c *gin.Context) {
	actorID, ok := middleware.UID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	role, ok := middleware.Role(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "工单 id 不合法")
		return
	}
	var req model.CreateCommentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.tickets.Comment(actorID, ticketID, role, req.Body)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}
