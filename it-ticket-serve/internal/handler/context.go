package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/middleware"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/service"
)

// actor 从 JWT 中间件取出当前用户。拿不到说明没过鉴权。
func actor(c *gin.Context) (uid int64, role string, ok bool) {
	uid, ok1 := middleware.UID(c)
	role, ok2 := middleware.Role(c)
	return uid, role, ok1 && ok2
}

func writeAPIError(c *gin.Context, err error) {
	var apiErr *service.Error
	if errors.As(err, &apiErr) {
		response.FailErr(c, apiErr.Status, apiErr.Code, apiErr.Name, apiErr.Message)
		return
	}
	logger.Error("未预期错误", "err", err)
	response.FailErr(c, http.StatusInternalServerError, response.CodeInternal, response.ErrInternal, "服务器内部错误")
}
