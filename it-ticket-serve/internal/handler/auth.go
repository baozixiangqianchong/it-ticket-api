package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	u, err := h.auth.Register(req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	out, err := h.auth.Login(req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.auth.Me(c.GetHeader("Authorization"))
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}
