package service

import (
	"database/sql"
	"errors"
	"strings"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

// AdminService 仅管理员使用的账号管理。路由上还会再挡一层 RequireAdmin。
type AdminService struct {
	users    *store.UserStore
	tickets  *store.TicketStore
	audits   *store.AuditStore
	notifs   *store.NotificationStore
	activity *store.ActivityStore
	invites  *store.InviteStore
}

func NewAdminService(
	users *store.UserStore,
	tickets *store.TicketStore,
	audits *store.AuditStore,
	notifs *store.NotificationStore,
	activity *store.ActivityStore,
	invites *store.InviteStore,
) *AdminService {
	return &AdminService{users: users, tickets: tickets, audits: audits, notifs: notifs, activity: activity, invites: invites}
}

func (s *AdminService) ListUsers(page, pageSize int, role, status, q string) (*model.UserPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	role = strings.TrimSpace(role)
	if role != "" && role != "user" && role != "agent" && role != "admin" {
		return nil, invalidArg("角色必须是 user / agent / admin")
	}
	status = strings.TrimSpace(status)
	if status != "" && status != model.UserActive && status != model.UserDisabled {
		return nil, invalidArg("状态必须是 active / disabled")
	}

	total, err := s.users.Count(role, status, q)
	if err != nil {
		return nil, err
	}
	list, err := s.users.List((page-1)*pageSize, pageSize, role, status, q)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(list))
	for _, u := range list {
		ids = append(ids, u.ID)
	}
	openCounts := map[int64]int64{}
	if s.tickets != nil {
		openCounts, err = s.tickets.CountOpenByAssignees(ids)
		if err != nil {
			return nil, err
		}
	}
	items := make([]model.AdminUser, 0, len(list))
	for _, u := range list {
		item := u.AdminPublic()
		item.OpenTicketCount = openCounts[u.ID]
		items = append(items, item)
	}
	return &model.UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) UpdateRole(actorID, userID int64, role string) (*model.AdminUser, error) {
	role = strings.TrimSpace(role)
	if role != "user" && role != "agent" && role != "admin" {
		return nil, invalidArg("角色必须是 user / agent / admin")
	}

	u, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("用户不存在")
		}
		return nil, err
	}

	// 不能拿掉自己的管理员：改成功后 JWT 下一请求就进不了 /admin，页面会误报「没有权限」。
	if actorID == userID && u.Role == "admin" && role != "admin" {
		return nil, permissionDenied("不能取消自己的管理员身份")
	}

	if u.Role == "admin" && role != "admin" {
		n, err := s.users.CountActiveByRole("admin")
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			return nil, lastAdmin()
		}
	}

	released := int64(0)
	if u.Role != role {
		from := u.Role
		n, err := s.applyAccountChange(actorID, userID, model.AccountActionRole, from, role, shouldReleaseTickets(from, role, false), "处理人改为"+roleCN(role)+"，未关单退回待派池", func(tx *sql.Tx) error {
			return s.users.UpdateRoleTx(tx, userID, role)
		})
		if err != nil {
			return nil, err
		}
		released = n
		logger.Info("用户角色已变更", "user_id", userID, "from", from, "to", role, "released", n)
		u.Role = role
		s.notifyAccount(userID, actorID, model.NotifyRole,
			"你的角色已改为"+roleCN(role),
			"管理员将你从"+roleCN(from)+"改为"+roleCN(role)+"，刷新后生效")
	}
	pub := u.AdminPublic()
	pub.ReleasedCount = released
	return &pub, nil
}

func (s *AdminService) UpdateStatus(actorID, userID int64, status string) (*model.AdminUser, error) {
	status = strings.TrimSpace(status)
	if status != model.UserActive && status != model.UserDisabled {
		return nil, invalidArg("状态必须是 active / disabled")
	}

	u, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("用户不存在")
		}
		return nil, err
	}
	if actorID == userID && status == model.UserDisabled {
		return nil, permissionDenied("不能停用自己的账号")
	}
	if u.Role == "admin" && status == model.UserDisabled && !u.Disabled() {
		n, err := s.users.CountActiveByRole("admin")
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			return nil, lastActiveAdmin()
		}
	}
	released := int64(0)
	if u.Status != status {
		from := u.Status
		n, err := s.applyAccountChange(actorID, userID, model.AccountActionStatus, from, status, shouldReleaseTickets(u.Role, u.Role, status == model.UserDisabled), "处理人已停用，未关单退回待派池", func(tx *sql.Tx) error {
			return s.users.UpdateStatusTx(tx, userID, status)
		})
		if err != nil {
			return nil, err
		}
		released = n
		logger.Info("用户状态已变更", "user_id", userID, "from", from, "to", status, "released", n)
		u.Status = status
		if status == model.UserActive {
			s.notifyAccount(userID, actorID, model.NotifyAccount, "账号已恢复使用", "管理员已重新启用你的账号")
		}
	}
	pub := u.AdminPublic()
	pub.ReleasedCount = released
	return &pub, nil
}

func (s *AdminService) applyAccountChange(actorID, userID int64, action, from, to string, release bool, releaseReason string, write func(*sql.Tx) error) (int64, error) {
	tx, err := s.users.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := write(tx); err != nil {
		return 0, err
	}
	if s.activity != nil {
		if err := s.activity.InsertAccount(tx, actorID, userID, action, &from, to); err != nil {
			return 0, err
		}
	}
	released := int64(0)
	if release {
		n, err := s.releaseOpenTickets(tx, actorID, userID, releaseReason)
		if err != nil {
			return 0, err
		}
		released = n
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return released, nil
}

func (s *AdminService) releaseOpenTickets(tx *sql.Tx, actorID, userID int64, reason string) (int64, error) {
	if s.tickets == nil || s.audits == nil {
		return 0, nil
	}
	list, err := s.tickets.ListOpenAssignedTx(tx, userID)
	if err != nil {
		return 0, err
	}
	reasonPtr := &reason
	for i := range list {
		t := &list[i]
		from := t.Status
		if err := s.tickets.ApplyTransition(tx, t.ID, store.TicketUpdate{
			Status:        model.StatusOpen,
			ClearAssignee: true,
			ClearClosedAt: true,
		}); err != nil {
			return 0, err
		}
		if err := s.audits.Insert(tx, t.ID, actorID, model.AuditRelease, &from, reasonPtr, model.StatusOpen); err != nil {
			return 0, err
		}
		if s.notifs != nil && t.CreatorID != actorID {
			id := t.ID
			if err := s.notifs.Insert(tx, t.CreatorID, &id, model.NotifyRelease, ticketSubject(t), clipRunes(reason, 500)); err != nil {
				return 0, err
			}
		}
	}
	return int64(len(list)), nil
}

func (s *AdminService) ListActivity(page, pageSize int, kind, action, q, fromDay, toDay string) (*model.ActivityPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	kind = strings.TrimSpace(kind)
	if kind != "" && kind != model.ActivityKindTicket && kind != model.ActivityKindAccount {
		return nil, invalidArg("kind 必须是 ticket / account")
	}
	action = strings.TrimSpace(action)
	if action != "" && !validActivityAction(action) {
		return nil, invalidArg("不支持的审计动作")
	}

	from, to, err := parseDayRange(fromDay, toDay)
	if err != nil {
		return nil, err
	}

	f := model.ActivityFilter{Kind: kind, Action: action, Q: q, From: from, To: to, Limit: pageSize, Offset: (page - 1) * pageSize}
	total, err := s.activity.Count(f)
	if err != nil {
		return nil, err
	}
	rows, err := s.activity.List(f)
	if err != nil {
		return nil, err
	}
	items := make([]model.PublicActivity, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Public())
	}
	return &model.ActivityPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func validActivityAction(action string) bool {
	switch action {
	case model.AuditCreate, model.AuditAssign, model.AuditClaim, model.AuditStart,
		model.AuditWait, model.AuditResume, model.AuditResolve, model.AuditClose,
		model.AuditReopen, model.AuditCancel, model.AuditTransfer, model.AuditEdit, model.AuditRelease,
		model.AccountActionRole, model.AccountActionStatus:
		return true
	default:
		return false
	}
}
