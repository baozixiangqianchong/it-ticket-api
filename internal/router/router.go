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
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestID(), middleware.AccessLog())

	r.GET("/healthz", handler.Healthz(db))

	users := store.NewUserStore(db)
	auth := handler.NewAuthHandler(service.NewAuthService(users, jwtSecret))
	tickets := handler.NewTicketHandler(service.NewTicketService(store.NewTicketStore(db), store.NewAuditStore(db), users, store.NewCommentStore(db)))
	admin := handler.NewAdminHandler(service.NewAdminService(users))

	v1 := r.Group("/api/v1")
	{
		// 注册、登录是公开的：还没有 token，不能走 JWT。
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)

		// 其余 /api/v1 接口都要登录。JWT 中间件验签失败会 401 并中止。
		authed := v1.Group("")
		authed.Use(middleware.JWT(jwtSecret))
		{
			authed.GET("/me", auth.Me)                            // GET /api/v1/me，用 token 换当前用户
			authed.POST("/tickets/create", tickets.Create)        // POST /api/v1/tickets，登录用户提单
			authed.GET("/tickets/list", tickets.List)             // GET /api/v1/tickets，登录用户查看工单
			authed.GET("/tickets/:id", tickets.GetDetail)         // GET /api/v1/tickets/:id，登录用户查看工单详情
			authed.POST("/tickets/:id/assign", tickets.Assign)    // POST /api/v1/tickets/:id/assign，登录用户指派工单
			authed.POST("/tickets/:id/update", tickets.Update)    // POST /api/v1/tickets/:id/update，按动作改状态
			authed.POST("/tickets/:id/comments", tickets.Comment) // POST /api/v1/tickets/:id/comments，对可见工单留言

			// /admin/* 再验一次 role==admin。员工、IT 到这里就是 403。
			adm := authed.Group("/admin")
			adm.Use(middleware.RequireAdmin())
			{
				adm.GET("/users", admin.ListUsers)            // GET /api/v1/admin/users
				adm.POST("/users/:id/role", admin.UpdateRole) // POST /api/v1/admin/users/:id/role
			}
		}
	}

	return r
}
