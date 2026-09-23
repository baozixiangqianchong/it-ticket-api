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
}

func NewTicketService(tickets *store.TicketStore, audits *store.AuditStore, users *store.UserStore, comments *store.CommentStore) *TicketService {
	return &TicketService{tickets: tickets, audits: audits, users: users, comments: comments}
}

func (s *TicketService) Create(creatorID int64, role string, in model.CreateTicketInput) (*model.PublicTicket, error) {
	title, description, category, err := validateCreateTicket(in)
	if err != nil {
		return nil, err
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id, err := s.tickets.Insert(tx, title, description, category, creatorID)
	if err != nil {
		return nil, err
	}
	if err := s.audits.Insert(tx, id, creatorID, model.AuditCreate, nil, nil, model.StatusOpen); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已创建", "ticket_id", id, "uid", creatorID, "category", category)
	return s.publicByID(id, creatorID, role)
}

func validateCreateTicket(in model.CreateTicketInput) (title, description, category string, err error) {
	title = strings.TrimSpace(in.Title)
	description = strings.TrimSpace(in.Description)
	category = strings.TrimSpace(in.Category)

	n := utf8.RuneCountInString(title)
	if n == 0 || n > 120 {
		return "", "", "", invalidArg("标题必填，最多 120 个字")
	}
	if description == "" {
		return "", "", "", invalidArg("描述必填")
	}
	if !validTicketCategory(category) {
		return "", "", "", invalidArg("分类必须是 hardware / software / network / other")
	}
	return title, description, category, nil
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

	// 员工没有待领 / 待我处理盒子，直接空页，不要 403。
	if role == "user" && (q.Scope == model.ScopePool || q.Scope == model.ScopeAssigned) {
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
		pub := rows[i].Public()
		pub.AvailableActions = availableActions(&rows[i], userID, role, now)
		items = append(items, pub)
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
	out.Assigned = assigned
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
	q.Q = strings.TrimSpace(q.Q)
	q.Scope = strings.TrimSpace(q.Scope)
	if q.Scope == "" {
		q.Scope = model.ScopeAll
	}
	switch q.Scope {
	case model.ScopeAll, model.ScopeCreated, model.ScopeAssigned, model.ScopePool:
	default:
		return q, invalidArg("scope 必须是 all / created / assigned / pool")
	}
	if q.Status != "" && !validTicketStatus(q.Status) {
		return q, invalidArg("非法状态")
	}
	if q.Category != "" && !validTicketCategory(q.Category) {
		return q, invalidArg("分类必须是 hardware / software / network / other")
	}
	return q, nil
}

func listFilter(userID int64, role string, q model.TicketListQuery) model.TicketFilter {
	f := model.TicketFilter{
		ViewerID:   userID,
		ViewerRole: role,
		Scope:      q.Scope,
		Status:     q.Status,
		Category:   q.Category,
		Q:          q.Q,
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

	pub := t.Public()
	pub.AvailableActions = availableActions(t, userID, role, time.Now())

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

	agent, err := s.users.FindByID(assigneeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invalidArg("受理人必须是 IT 处理人")
		}
		return nil, err
	}
	if agent.Role != "agent" {
		return nil, invalidArg("受理人必须是 IT 处理人")
	}

	t, err := s.tickets.FindByID(ticketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	switch t.Status {
	case model.StatusOpen, model.StatusAssigned, model.StatusInProgress:
	default:
		return nil, ticketInvalidTransition("当前状态不能指派")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.tickets.UpdateAssignee(tx, ticketID, assigneeID); err != nil {
		return nil, err
	}
	from := t.Status
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditAssign, &from, nil, model.StatusAssigned); err != nil {
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
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditClaim, &from, nil, model.StatusAssigned); err != nil {
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
		return nil, invalidArg("必须指定动作：start / resolve / close / reopen / cancel")
	}

	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}

	result, err := resolveTicketAction(t, actorID, role, action, reason, time.Now())
	if err != nil {
		return nil, err
	}

	// 重开时原处理人必须还是在职 IT；否则退回待派池，避免指给一个已经不是 agent 的人。
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
	if u.Role != "agent" {
		return false, "原处理人不再是 IT"
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

	c, err := s.comments.Insert(ticketID, actorID, body)
	if err != nil {
		return nil, err
	}
	logger.Info("工单已评论", "ticket_id", ticketID, "uid", actorID, "comment_id", c.ID)
	pub := c.Public()
	return &pub, nil
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
	pub := t.Public()
	if role != "" {
		pub.AvailableActions = availableActions(t, viewerID, role, time.Now())
	}
	return &pub, nil
}
