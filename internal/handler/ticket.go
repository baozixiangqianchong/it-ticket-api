package handler

import (
	"net/http"

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
