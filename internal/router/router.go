package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/handler"
	"it-ticket-api/internal/middleware"
	"it-ticket-api/internal/service"
	"it-ticket-api/internal/store"
)

func New(db *sql.DB, jwtSecret string) *gin.Engine {
	r := gin.New()
	// Recovery：handler panic 时返回 500，进程不崩。
	// RequestID：每条请求一个 ID。AccessLog：结束后打印路径、参数、状态码。
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.AccessLog())

	r.GET("/healthz", handler.Healthz(db))

	auth := handler.NewAuthHandler(service.NewAuthService(store.NewUserStore(db), jwtSecret))
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)
	}

	return r
}
