package router

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/handler"
	"it-ticket-api/internal/middleware"
	"it-ticket-api/internal/service"
	"it-ticket-api/internal/store"
)

func New(db *sql.DB, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestID(), middleware.AccessLog())

	r.GET("/healthz", handler.Healthz(db))

	users := store.NewUserStore(db)
	auth := handler.NewAuthHandler(service.NewAuthService(users, jwtSecret))
	tickets := handler.NewTicketHandler(service.NewTicketService(store.NewTicketStore(db), store.NewAuditStore(db), users, store.NewCommentStore(db)))
	admin := handler.NewAdminHandler(service.NewAdminService(users))

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)

		authed := v1.Group("")
		// 角色以库为准：token 只证明 uid。
		authed.Use(middleware.JWT(jwtSecret, func(uid int64) (string, bool, error) {
			u, err := users.FindByID(uid)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return "", false, nil
				}
				return "", false, err
			}
			return u.Role, true, nil
		}))
		{
			authed.GET("/me", auth.Me)
			authed.POST("/tickets/create", tickets.Create)
			authed.GET("/tickets/list", tickets.List)
			authed.GET("/tickets/stats", tickets.Stats) // 必须在 /tickets/:id 前面
			authed.GET("/tickets/:id", tickets.GetDetail)
			authed.POST("/tickets/:id/assign", tickets.Assign)
			authed.POST("/tickets/:id/claim", tickets.Claim)
			authed.POST("/tickets/:id/update", tickets.Update)
			authed.POST("/tickets/:id/comments", tickets.Comment)

			adm := authed.Group("/admin")
			adm.Use(middleware.RequireAdmin())
			{
				adm.GET("/users", admin.ListUsers)
				adm.POST("/users/:id/role", admin.UpdateRole)
			}
		}
	}

	return r
}
