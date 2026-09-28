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
	notifs := store.NewNotificationStore(db)
	invites := store.NewInviteStore(db)
	ticketStore := store.NewTicketStore(db)
	auditStore := store.NewAuditStore(db)
	catalogStore := store.NewCatalogStore(db)
	catalog := service.NewCatalogService(catalogStore)
	auth := handler.NewAuthHandler(service.NewAuthService(users, notifs, invites, jwtSecret))
	tickets := handler.NewTicketHandler(service.NewTicketService(ticketStore, auditStore, users, store.NewCommentStore(db), notifs), catalog)
	admin := handler.NewAdminHandler(service.NewAdminService(users, ticketStore, auditStore, notifs, store.NewActivityStore(db), invites), catalog)
	notify := handler.NewNotificationHandler(service.NewNotificationService(notifs))

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)

		authed := v1.Group("")
		// 角色以库为准：token 只证明 uid。停用账号下一请求 401。
		authed.Use(middleware.JWT(jwtSecret, func(uid int64) (string, bool, bool, error) {
			u, err := users.FindByID(uid)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return "", false, false, nil
				}
				return "", false, false, err
			}
			return u.Role, u.Disabled(), true, nil
		}))
		{
			authed.GET("/me", auth.Me)
			authed.POST("/me/profile", auth.UpdateProfile)
			authed.POST("/me/password", auth.UpdatePassword)
			authed.GET("/agents", tickets.ListAgents)
			authed.GET("/ticket-templates", tickets.ListTemplates)
			authed.GET("/canned-replies", tickets.ListCannedReplies)
			authed.GET("/notifications", notify.List)
			authed.POST("/notifications/read-all", notify.ReadAll)
			authed.POST("/notifications/:id/read", notify.Read)
			authed.POST("/tickets/create", tickets.Create)
			authed.GET("/tickets/list", tickets.List)
			authed.GET("/tickets/stats", tickets.Stats) // 必须在 /tickets/:id 前面
			authed.GET("/tickets/board", tickets.Board)
			authed.POST("/tickets/close-stale", tickets.CloseStale)
			authed.GET("/tickets/:id", tickets.GetDetail)
			authed.POST("/tickets/:id/edit", tickets.Edit)
			authed.POST("/tickets/:id/assign", tickets.Assign)
			authed.POST("/tickets/:id/claim", tickets.Claim)
			authed.POST("/tickets/:id/transfer", tickets.Transfer)
			authed.POST("/tickets/:id/update", tickets.Update)
			authed.POST("/tickets/:id/comments", tickets.Comment)

			adm := authed.Group("/admin")
			adm.Use(middleware.RequireAdmin())
			{
				adm.GET("/users", admin.ListUsers)
				adm.GET("/audits", admin.ListAudits)
				adm.POST("/users/:id/role", admin.UpdateRole)
				adm.POST("/users/:id/status", admin.UpdateStatus)
				adm.GET("/invites", admin.ListInvites)
				adm.POST("/invites", admin.CreateInvite)
				adm.GET("/ticket-templates", admin.ListTemplates)
				adm.POST("/ticket-templates", admin.CreateTemplate)
				adm.POST("/ticket-templates/:id", admin.UpdateTemplate)
				adm.POST("/ticket-templates/:id/delete", admin.DeleteTemplate)
				adm.GET("/canned-replies", admin.ListReplies)
				adm.POST("/canned-replies", admin.CreateReply)
				adm.POST("/canned-replies/:id", admin.UpdateReply)
				adm.POST("/canned-replies/:id/delete", admin.DeleteReply)
			}
		}
	}

	return r
}
