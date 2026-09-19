package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/response"
)

// Healthz 健康检查：进程是否活着，以及 MySQL 能否 ping。
func Healthz(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			response.Fail(c, http.StatusServiceUnavailable, response.CodeUnavailable, "数据库不可用")
			return
		}
		response.OK(c, gin.H{"mysql": "ok"})
	}
}
