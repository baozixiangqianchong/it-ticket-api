package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

type NotificationHandler struct {
	notifs *service.NotificationService
}

func NewNotificationHandler(notifs *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifs: notifs}
}

func (h *NotificationHandler) List(c *gin.Context) {
	uid, _, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	page, pageSize := parsePage(c)
	unreadOnly := c.Query("unread") == "1" || c.Query("unread") == "true"
	out, err := h.notifs.List(uid, unreadOnly, page, pageSize)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *NotificationHandler) Read(c *gin.Context) {
	uid, _, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "通知 id 不合法")
		return
	}
	if err := h.notifs.MarkRead(uid, id); err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *NotificationHandler) ReadAll(c *gin.Context) {
	uid, _, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	if err := h.notifs.MarkAllRead(uid); err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, gin.H{})
}
