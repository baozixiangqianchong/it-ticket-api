package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/response"
)

// RequireAdmin 挂在 /api/v1/admin/* 上。前面必须已经过 JWT。
// 非 admin 直接 403 并中止，不要让请求进到用户列表或改角色。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := Role(c)
		if !ok || role != "admin" {
			response.Fail(c, http.StatusForbidden, response.CodeForbidden, "没有管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
