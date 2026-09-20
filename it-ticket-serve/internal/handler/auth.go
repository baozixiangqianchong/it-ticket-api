package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

// AuthHandler 处理注册、登录、当前用户的 HTTP 请求。
// 这里只负责：读参数、调 service、写响应。不校验密码规则、不拼 SQL、不解析 JWT。
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register POST /api/v1/auth/register
// 把 body 解成 RegisterInput，交给 service 建用户，成功则把公开资料放进统一信封返回。
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterInput
	// ShouldBindJSON：按 json 标签填结构体。末尾多逗号、缺引号等都不是合法 JSON。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidArg, "JSON 格式错误")
		return
	}
	u, err := h.auth.Register(req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u) // HTTP 200，code=0，data 是用户信息（无密码）
}

// Login POST /api/v1/auth/login
// 成功时 data 里是 token + 用户信息，给客户端以后放进 Authorization: Bearer <token>。
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

// Me GET /api/v1/me
// 路由已挂 JWT 中间件，能走到这里说明头里有合法 token。
// 把 Authorization 原样交给 service：再验一次、按 uid 查库、去掉密码。
func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.auth.Me(c.GetHeader("Authorization"))
	if err != nil {
		writeAPIError(c, err)
		return
	}
	response.OK(c, u)
}

// writeAPIError 把 service 的错误翻成 HTTP。
// service.Error：按里面的 Status/Code/中文 Message 原样返回。
// 其它错误：当内部故障，日志记详情，响应不暴露 SQL。
func writeAPIError(c *gin.Context, err error) {
	var apiErr *service.Error
	if errors.As(err, &apiErr) {
		response.Fail(c, apiErr.Status, apiErr.Code, apiErr.Message)
		return
	}
	log.Printf("internal error: %v", err)
	response.Fail(c, http.StatusInternalServerError, response.CodeInternal, "服务器内部错误")
}
