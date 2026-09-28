package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

// TicketService 工单业务。状态机、权限、工单与审计同一事务都放这里，不拼 SQL。
type TicketService struct {
	tickets  *store.TicketStore
	audits   *store.AuditStore
	users    *store.UserStore
	comments *store.CommentStore
	notifs   *store.NotificationStore
}

func NewTicketService(tickets *store.TicketStore, audits *store.AuditStore, users *store.UserStore, comments *store.CommentStore, notifs *store.NotificationStore) *TicketService {
	return &TicketService{tickets: tickets, audits: audits, users: users, comments: comments, notifs: notifs}
}

func (s *TicketService) Create(creatorID int64, role string, in model.CreateTicketInput) (*model.PublicTicket, error) {
	title, description, category, priority, err := validateCreateTicket(in, role)
	if err != nil {
		return nil, err
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id, err := s.tickets.Insert(tx, title, description, category, priority, creatorID)
	if err != nil {
		return nil, err
	}
	if err := s.audits.Insert(tx, id, creatorID, model.AuditCreate, nil, nil, model.StatusOpen); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已创建", "ticket_id", id, "uid", creatorID, "category", category, "priority", priority)
	return s.publicByID(id, creatorID, role)
}

func validateCreateTicket(in model.CreateTicketInput, role string) (title, description, category, priority string, err error) {
	title = strings.TrimSpace(in.Title)
	description = strings.TrimSpace(in.Description)
	category = strings.TrimSpace(in.Category)
	priority = strings.TrimSpace(in.Priority)
	if priority == "" {
		priority = model.PriorityP2
	}

	n := utf8.RuneCountInString(title)
	if n == 0 || n > 120 {
		return "", "", "", "", invalidArg("标题必填，最多 120 个字")
	}
	if description == "" {
		return "", "", "", "", invalidArg("描述必填")
	}
	if !validTicketCategory(category) {
		return "", "", "", "", invalidArg("分类必须是 hardware / software / network / other")
	}
	if !validTicketPriority(priority) {
		return "", "", "", "", invalidArg("优先级必须是 p1 / p2 / p3")
	}
	if err := denyP1(role, priority); err != nil {
		return "", "", "", "", err
	}
	return title, description, category, priority, nil
}

func validTicketPriority(s string) bool {
	switch s {
	case model.PriorityP1, model.PriorityP2, model.PriorityP3:
		return true
	default:
		return false
	}
}

func validTicketCategory(s string) bool {
	switch s {
	case model.CategoryHardware, model.CategorySoftware, model.CategoryNetwork, model.CategoryOther:
		return true
	default:
		return false
	}
}

func (s *TicketService) List(userID int64, role string, q model.TicketListQuery) (*model.TicketPage, error) {
	q, err := normalizeListQuery(q)
	if err != nil {
		return nil, err
	}

	// 员工没有待领 / 待我处理 / 等对方盒子，直接空页，不要 403。
	if role == "user" && (q.Scope == model.ScopePool || q.Scope == model.ScopeAssigned || q.Scope == model.ScopeWaiting) {
		return &model.TicketPage{Items: []model.PublicTicket{}, Total: 0, Page: q.Page, PageSize: q.PageSize}, nil
	}

	f := listFilter(userID, role, q)
	total, err := s.tickets.Count(f)
	if err != nil {
		return nil, err
	}
	rows, err := s.tickets.List(f)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]model.PublicTicket, 0, len(rows))
	for i := range rows {
		items = append(items, decorateTicket(&rows[i], userID, role, now))
	}
	return &model.TicketPage{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *TicketService) Stats(userID int64, role string) (*model.TicketStats, error) {
	created, err := s.tickets.Count(model.TicketFilter{ViewerID: userID, ViewerRole: role, Scope: model.ScopeCreated})
	if err != nil {
		return nil, err
	}
	all, err := s.tickets.Count(model.TicketFilter{ViewerID: userID, ViewerRole: role, Scope: model.ScopeAll})
	if err != nil {
		return nil, err
	}

	out := &model.TicketStats{Created: created, All: all}
	if role == "user" {
		return out, nil
	}
	assigned, err := s.tickets.Count(model.TicketFilter{ViewerID: userID, ViewerRole: role, Scope: model.ScopeAssigned})
	if err != nil {
		return nil, err
	}
	pool, err := s.tickets.Count(model.TicketFilter{ViewerID: userID, ViewerRole: role, Scope: model.ScopePool})
	if err != nil {
		return nil, err
	}
	waiting, err := s.tickets.Count(model.TicketFilter{ViewerID: userID, ViewerRole: role, Scope: model.ScopeWaiting})
	if err != nil {
		return nil, err
	}
	out.Assigned = assigned
	out.Waiting = waiting
	out.Pool = pool
	return out, nil
}

func normalizeListQuery(q model.TicketListQuery) (model.TicketListQuery, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 50 {
		q.PageSize = 50
	}
	q.Status = strings.TrimSpace(q.Status)
	q.Category = strings.TrimSpace(q.Category)
	q.Priority = strings.TrimSpace(q.Priority)
	q.Q = strings.TrimSpace(q.Q)
	q.Scope = strings.TrimSpace(q.Scope)
	if q.Scope == "" {
		q.Scope = model.ScopeAll
	}
	switch q.Scope {
	case model.ScopeAll, model.ScopeCreated, model.ScopeAssigned, model.ScopeWaiting, model.ScopePool:
	default:
		return q, invalidArg("scope 必须是 all / created / assigned / waiting / pool")
	}
	if q.Status != "" && !validTicketStatus(q.Status) {
		return q, invalidArg("非法状态")
	}
	if q.Category != "" && !validTicketCategory(q.Category) {
		return q, invalidArg("分类必须是 hardware / software / network / other")
	}
	if q.Priority != "" && !validTicketPriority(q.Priority) {
		return q, invalidArg("优先级必须是 p1 / p2 / p3")
	}
	q.TimeField = strings.TrimSpace(q.TimeField)
	if !validTimeField(q.TimeField) {
		return q, invalidArg("time_field 必须是 created_at / updated_at / closed_at")
	}
	if _, _, err := parseDayRange(q.FromDay, q.ToDay); err != nil {
		return q, err
	}
	return q, nil
}

func listFilter(userID int64, role string, q model.TicketListQuery) model.TicketFilter {
	from, to, _ := parseDayRange(q.FromDay, q.ToDay)
	f := model.TicketFilter{
		ViewerID:   userID,
		ViewerRole: role,
		Scope:      q.Scope,
		Status:     q.Status,
		Category:   q.Category,
		Priority:   q.Priority,
		Q:          q.Q,
		TimeField:  q.TimeField,
		From:       from,
		To:         to,
		Limit:      q.PageSize,
		Offset:     (q.Page - 1) * q.PageSize,
	}
	// 按处理人筛只有管理员有意义，其他人传入直接丢掉。
	if role == "admin" {
		f.AssigneeID = q.AssigneeID
	}
	return f
}

func (s *TicketService) GetDetail(userID, id int64, role string) (*model.TicketDetail, error) {
	t, err := s.visibleTicket(userID, id, role)
	if err != nil {
		return nil, err
	}
	comments, err := s.comments.ListByTicketID(id)
	if err != nil {
		return nil, err
	}
	audits, err := s.audits.ListByTicketID(id)
	if err != nil {
		return nil, err
	}

	pub := decorateTicket(t, userID, role, time.Now())

	commentOut := make([]model.PublicComment, 0, len(comments))
	for _, c := range comments {
		commentOut = append(commentOut, c.Public())
	}
	auditOut := make([]model.PublicAudit, 0, len(audits))
	for _, a := range audits {
		auditOut = append(auditOut, a.Public())
	}
	return &model.TicketDetail{PublicTicket: pub, Comments: commentOut, Audits: auditOut}, nil
}

func (s *TicketService) Assign(actorID, ticketID int64, role string, assigneeID int64) (*model.PublicTicket, error) {
	if role != "admin" {
		return nil, permissionDenied("没有权限指派工单")
	}
	if assigneeID <= 0 {
		return nil, invalidArg("必须指定受理人")
	}
	if _, err := s.requireActiveAgent(assigneeID); err != nil {
		return nil, err
	}

	t, err := s.tickets.FindByID(ticketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	switch t.Status {
	case model.StatusOpen, model.StatusAssigned, model.StatusInProgress, model.StatusPending:
	default:
		return nil, ticketInvalidTransition("当前状态不能指派")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	from := t.Status
	to := statusAfterHandoff(from)
	if err := s.tickets.UpdateAssignee(tx, ticketID, assigneeID, to); err != nil {
		return nil, err
	}
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditAssign, &from, nil, to); err != nil {
		return nil, err
	}
	name := displayName(s.users, actorID)
	if err := s.notifyTicket(tx, assigneeID, actorID, t, model.NotifyAssign, name+" 已指派给你"); err != nil {
		return nil, err
	}
	if t.AssigneeID != nil && *t.AssigneeID != assigneeID {
		if err := s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyAssign, name+" 已把这张单改派给他人"); err != nil {
			return nil, err
		}
	}
	if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyAssign, name+" 已指派处理人"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已指派", "ticket_id", ticketID, "actor_id", actorID, "assignee_id", assigneeID, "from", from)
	return s.publicByID(ticketID, actorID, role)
}

func (s *TicketService) Claim(actorID, ticketID int64, role string) (*model.PublicTicket, error) {
	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}
	if role != "agent" && role != "admin" {
		return nil, permissionDenied("只有 IT 或管理员可以领取工单")
	}
	if t.Status != model.StatusOpen {
		return nil, ticketInvalidTransition("当前状态不能领取")
	}
	if t.AssigneeID != nil {
		return nil, ticketAlreadyAssigned("工单已被领取")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ok, err := s.tickets.ClaimOpen(tx, ticketID, actorID)
	if err != nil {
		return nil, err
	}
	if !ok {
		logger.Warn("工单自领冲突", "ticket_id", ticketID, "uid", actorID)
		return nil, ticketAlreadyAssigned("工单已被领取")
	}
	from := model.StatusOpen
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditClaim, &from, nil, model.StatusInProgress); err != nil {
		return nil, err
	}
	if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyClaim, displayName(s.users, actorID)+" 已领取这张单"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已自领", "ticket_id", ticketID, "uid", actorID)
	return s.publicByID(ticketID, actorID, role)
}

func (s *TicketService) Update(actorID, ticketID int64, role, action, reason string) (*model.PublicTicket, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, invalidArg("必须指定动作：start / resolve / close / reopen / cancel / wait / resume")
	}

	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}

	result, err := resolveTicketAction(t, actorID, role, action, reason, time.Now())
	if err != nil {
		return nil, err
	}

	// 重开时原处理人必须还是在职 IT 或管理员；否则退回待派池。
	if action == model.ActionReopen {
		if keep, why := s.keepAssigneeOnReopen(t); !keep {
			result.To = model.StatusOpen
			result.ClearAssignee = true
			logger.Info("重开退回待派池", "ticket_id", ticketID, "reason", why)
		}
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.tickets.ApplyTransition(tx, ticketID, store.TicketUpdate{
		Status:        result.To,
		ClearAssignee: result.ClearAssignee,
		SetClosedNow:  result.SetClosedAt,
		ClearClosedAt: result.ClearClosedAt,
	}); err != nil {
		return nil, err
	}
	from := t.Status
	if err := s.audits.Insert(tx, ticketID, actorID, action, &from, reasonPtr(reason), result.To); err != nil {
		return nil, err
	}
	if err := s.notifyAfterUpdate(tx, t, actorID, action, reason, result); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单状态变更", "ticket_id", ticketID, "uid", actorID, "action", action, "from", from, "to", result.To)
	return s.publicByID(ticketID, actorID, role)
}

func (s *TicketService) keepAssigneeOnReopen(t *model.Ticket) (bool, string) {
	if t.AssigneeID == nil {
		return false, "原单没有处理人"
	}
	u, err := s.users.FindByID(*t.AssigneeID)
	if err != nil {
		return false, "原处理人已不存在"
	}
	if !canHandleTickets(u.Role) {
		return false, "原处理人不再能接单"
	}
	if u.Disabled() {
		return false, "原处理人已停用"
	}
	return true, ""
}

func (s *TicketService) Comment(actorID, ticketID int64, role, body string) (*model.PublicComment, error) {
	body = strings.TrimSpace(body)
	n := utf8.RuneCountInString(body)
	if n == 0 || n > 2000 {
		return nil, invalidArg("评论必填，最多 2000 个字")
	}

	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}
	if t.Status == model.StatusClosed {
		return nil, ticketClosed("工单已关闭，不能再评论")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	commentID, err := s.comments.InsertTx(tx, ticketID, actorID, body)
	if err != nil {
		return nil, err
	}

	// 提单人在 pending 下评论，同一事务拉回处理中。
	if t.Status == model.StatusPending && t.CreatorID == actorID {
		if err := s.tickets.ApplyTransition(tx, ticketID, store.TicketUpdate{Status: model.StatusInProgress}); err != nil {
			return nil, err
		}
		from := model.StatusPending
		if err := s.audits.Insert(tx, ticketID, actorID, model.AuditResume, &from, nil, model.StatusInProgress); err != nil {
			return nil, err
		}
		logger.Info("评论拉回处理中", "ticket_id", ticketID, "uid", actorID)
	}

	excerpt := displayName(s.users, actorID) + "：" + clipRunes(body, 80)
	if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyComment, excerpt); err != nil {
		return nil, err
	}
	if t.AssigneeID != nil {
		if err := s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyComment, excerpt); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	c, err := s.comments.FindByID(commentID)
	if err != nil {
		return nil, err
	}
	logger.Info("工单已评论", "ticket_id", ticketID, "uid", actorID, "comment_id", c.ID)
	pub := c.Public()
	return &pub, nil
}

func (s *TicketService) Transfer(actorID, ticketID int64, role string, assigneeID int64, reason string) (*model.PublicTicket, error) {
	if role != "admin" && role != "agent" {
		return nil, permissionDenied("只有当前处理人或管理员可以转派")
	}
	if assigneeID <= 0 {
		return nil, invalidArg("必须指定受理人")
	}
	if assigneeID == actorID {
		return nil, invalidArg("不能转派给自己")
	}
	if _, err := s.requireActiveAgent(assigneeID); err != nil {
		return nil, err
	}

	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}
	isAssignee := t.AssigneeID != nil && *t.AssigneeID == actorID
	if role != "admin" && !isAssignee {
		return nil, permissionDenied("只有当前处理人或管理员可以转派")
	}
	switch t.Status {
	case model.StatusAssigned, model.StatusInProgress, model.StatusPending:
	default:
		return nil, ticketInvalidTransition("当前状态不能转派")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	from := t.Status
	to := statusAfterHandoff(from)
	if err := s.tickets.UpdateAssignee(tx, ticketID, assigneeID, to); err != nil {
		return nil, err
	}
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditTransfer, &from, reasonPtr(reason), to); err != nil {
		return nil, err
	}
	name := displayName(s.users, actorID)
	if err := s.notifyTicket(tx, assigneeID, actorID, t, model.NotifyTransfer, withReason(name+" 转派给你", reason)); err != nil {
		return nil, err
	}
	if t.AssigneeID != nil {
		if err := s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyTransfer, name+" 已把这张单转走"); err != nil {
			return nil, err
		}
	}
	if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyTransfer, name+" 已将工单转派"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已转派", "ticket_id", ticketID, "actor_id", actorID, "assignee_id", assigneeID, "from", from)
	return s.publicByID(ticketID, actorID, role)
}

func (s *TicketService) ListAgents() (*model.AgentList, error) {
	list, err := s.users.ListActiveHandlers()
	if err != nil {
		return nil, err
	}
	items := make([]model.AgentRef, 0, len(list))
	for _, u := range list {
		items = append(items, model.AgentRef{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName})
	}
	return &model.AgentList{Items: items}, nil
}

func (s *TicketService) requireActiveAgent(id int64) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invalidArg("受理人必须是在职 IT 或管理员")
		}
		return nil, err
	}
	if !canHandleTickets(u.Role) {
		return nil, invalidArg("受理人必须是在职 IT 或管理员")
	}
	if u.Disabled() {
		return nil, invalidArg("受理人账号已停用")
	}
	return u, nil
}

func (s *TicketService) notifyAfterUpdate(tx *sql.Tx, t *model.Ticket, actorID int64, action, reason string, result actionResult) error {
	name := displayName(s.users, actorID)
	switch action {
	case model.ActionStart:
		return s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyStart, name+" 已开始处理")
	case model.ActionWait:
		return s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyWait, withReason(name+" 等你补充信息", reason))
	case model.ActionResume:
		return s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyResume, name+" 已继续处理")
	case model.ActionResolve:
		return s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyResolve, name+" 已标记为已解决，请确认后关闭")
	case model.ActionCancel:
		if t.AssigneeID != nil {
			return s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyCancel, name+" 已撤回工单")
		}
	case model.ActionClose:
		if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyClose, name+" 已关闭工单"); err != nil {
			return err
		}
		if t.AssigneeID != nil {
			return s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyClose, name+" 已关闭工单")
		}
	case model.ActionReopen:
		body := withReason(name+" 已重开", reason)
		if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyReopen, body); err != nil {
			return err
		}
		if result.To == model.StatusAssigned && t.AssigneeID != nil {
			return s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyReopen, body)
		}
	}
	return nil
}

func (s *TicketService) visibleTicket(userID, ticketID int64, role string) (*model.Ticket, error) {
	t, err := s.tickets.GetByViewer(userID, ticketID, role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	return t, nil
}

func (s *TicketService) publicByID(id, viewerID int64, role string) (*model.PublicTicket, error) {
	t, err := s.tickets.FindByID(id)
	if err != nil {
		return nil, err
	}
	pub := decorateTicket(t, viewerID, role, time.Now())
	return &pub, nil
}

func decorateTicket(t *model.Ticket, viewerID int64, role string, now time.Time) model.PublicTicket {
	pub := t.Public()
	if role != "" {
		pub.AvailableActions = availableActions(t, viewerID, role, now)
	}
	pub.Stale = ticketStale(t, now)
	return pub
}
