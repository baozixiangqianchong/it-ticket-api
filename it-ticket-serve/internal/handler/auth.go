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

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	uid, _, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	var req model.UpdateProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	u, err := h.auth.UpdateProfile(uid, req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	uid, _, ok := actor(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
		return
	}
	var req model.UpdatePasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	if err := h.auth.UpdatePassword(uid, req); err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, gin.H{})
}
