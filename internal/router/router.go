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
	tickets := handler.NewTicketHandler(service.NewTicketService(store.NewTicketStore(db), store.NewAuditStore(db)))

	v1 := r.Group("/api/v1")
	{
		// 注册、登录是公开的：还没有 token，不能走 JWT。
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)

		// 其余 /api/v1 接口都要登录。JWT 中间件验签失败会 401 并中止。
		authed := v1.Group("")
		authed.Use(middleware.JWT(jwtSecret))
		{
			authed.GET("/me", auth.Me)              // GET /api/v1/me，用 token 换当前用户
			authed.POST("/tickets", tickets.Create) // POST /api/v1/tickets，登录用户提单
		}
	}

	return r
}
