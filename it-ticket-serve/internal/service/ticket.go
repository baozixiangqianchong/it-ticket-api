package service

import (
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

// TicketService 工单业务。状态机、权限、工单与审计同一事务都放这里，不拼 SQL。
type TicketService struct {
	tickets  *store.TicketStore  // 写/读 tickets 表
	audits   *store.AuditStore   // 写 audit_logs 表
	users    *store.UserStore    // 指派时校验目标是不是 agent
	comments *store.CommentStore // 写 ticket_comments 表
}

func NewTicketService(tickets *store.TicketStore, audits *store.AuditStore, users *store.UserStore, comments *store.CommentStore) *TicketService {
	return &TicketService{tickets: tickets, audits: audits, users: users, comments: comments}
}

// Create 员工（以及 agent / admin 以员工身份）提单。
// 服务端写死 status=open、assignee 为空；客户端就算传这两个字段也没用。
// 工单行和审计行同一事务：只成功一半会回滚，不会出现「有单无审计」。
func (s *TicketService) Create(creatorID int64, in model.CreateTicketInput) (*model.PublicTicket, error) {
	title, description, category, err := validateCreateTicket(in)
	if err != nil {
		return nil, err
	}

	// 事务必须从这里开：store 只执行 SQL，不决定「这两条要不要绑在一起」。
	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	// defer Rollback：后面任何 return（校验后的 DB 错误）都会撤掉未提交的写入。
	// Commit 成功后 Rollback 是空操作，不会把已经提交的再撤掉。
	defer tx.Rollback()

	id, err := s.tickets.Insert(tx, title, description, category, creatorID)
	if err != nil {
		return nil, err
	}
	// from_status 传 nil：创建没有「原状态」。to_status=open，action=create。
	if err := s.audits.Insert(tx, id, creatorID, model.AuditCreate, nil, model.StatusOpen); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 提交后再按 id 读一遍，才能带上库生成的 created_at / updated_at。
	t, err := s.tickets.FindByID(id)
	if err != nil {
		return nil, err
	}
	// Public() 只返回接口该看的字段（和 User.Public 一个思路）。
	pub := t.Public()
	return &pub, nil
}

// validateCreateTicket 整理并校验入参。空白标题、空描述、非法分类都是 400。
func validateCreateTicket(in model.CreateTicketInput) (title, description, category string, err error) {
	title = strings.TrimSpace(in.Title)
	description = strings.TrimSpace(in.Description)
	category = strings.TrimSpace(in.Category)

	// 标题按「字」计数，中文「显示器不亮」是 5，和库里 VARCHAR(120) 对齐。
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

// validTicketCategory 只放行文档里的四个枚举。其它字符串不要等 MySQL ENUM 报错再失败。
func validTicketCategory(s string) bool {
	switch s {
	case model.CategoryHardware, model.CategorySoftware, model.CategoryNetwork, model.CategoryOther:
		return true
	default:
		return false
	}
}

// List 同一条列表：按角色缩小可见范围，不给三种人各开一个接口。
func (s *TicketService) List(userID int64, role string) ([]model.PublicTicket, error) {
	return s.tickets.ListByViewer(userID, role)
}

// GetDetail 工单详情：可见范围和列表相同。看不见或没有这张单，都是 404。
func (s *TicketService) GetDetail(userID, id int64, role string) (*model.TicketDetail, error) {
	t, err := s.tickets.GetDetailByViewer(userID, id, role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	comments, err := s.comments.ListByTicketID(id)
	if err != nil {
		return nil, err
	}
	pubs := make([]model.PublicComment, 0, len(comments))
	for _, c := range comments {
		// ListByTicketID 已经按 ticket_id 查了。这里再挡一层：
		if c.TicketID != id {
			continue
		}
		pubs = append(pubs, c.Public())
	}
	return &model.TicketDetail{PublicTicket: t.Public(), Comments: pubs}, nil
}

// Assign 仅 admin 把工单派给某位 agent。员工和 IT 都不能指派。
// actorID 是当前管理员，只写进审计；assigneeID 才写入 tickets.assignee_id。
func (s *TicketService) Assign(actorID, ticketID int64, role string, assigneeID int64) (*model.PublicTicket, error) {
	// 已登录但角色不够：403。员工、IT 走这里，不要用 401。
	if role != "admin" {
		return nil, permissionDenied("没有权限指派工单")
	}
	// 没传或传了 0：JSON 里缺 assignee_id，还没查库。
	if assigneeID <= 0 {
		return nil, invalidArg("必须指定受理人")
	}

	// 目标必须是库里存在的 agent，不能派给员工或另一个管理员。
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
	// 待派、已派、处理中可以指派或改派；已解决 / 已关闭不行。
	switch t.Status {
	case model.StatusOpen, model.StatusAssigned, model.StatusInProgress:
	default:
		return nil, conflict("当前状态不能指派")
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	// 改工单和写审计必须一起成功；中途失败 Rollback 会两行都撤掉。
	defer tx.Rollback()

	if err := s.tickets.UpdateAssignee(tx, ticketID, assigneeID); err != nil {
		return nil, err
	}
	// from 用改之前的状态；actorID 记「谁派的」，不要写成受理人。
	from := t.Status
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditAssign, &from, model.StatusAssigned); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 提交后再读一遍，返回带上新处理人和 assigned。
	updated, err := s.tickets.FindByID(ticketID)
	if err != nil {
		return nil, err
	}
	pub := updated.Public()
	return &pub, nil
}

// Update 按动作推进状态。三种角色都能改，但不是每个人都能做每一个动作。
// 客户端传 action，不传 status，避免绕过状态机。
func (s *TicketService) Update(actorID, ticketID int64, role, action string) (*model.PublicTicket, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, invalidArg("必须指定动作：start / resolve / close / reopen")
	}

	// 看不见这张单和没有这张单一样，404。
	t, err := s.tickets.GetDetailByViewer(actorID, ticketID, role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}

	// 判断这一步能不能走、走到哪。非法跳转 409，人不对 403。
	to, clearAssignee, err := resolveTicketAction(t, actorID, role, action)
	if err != nil {
		return nil, err
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 更新工单状态
	if err := s.tickets.UpdateStatus(tx, ticketID, to, clearAssignee); err != nil {
		return nil, err
	}
	from := t.Status
	// 写审计
	// from_status 用改之前的状态；actorID 记「谁做的」，不要写成处理人。
	if err := s.audits.Insert(tx, ticketID, actorID, action, &from, to); err != nil {
		return nil, err
	}
	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 提交后再读一遍，返回带上新状态和处理人
	updated, err := s.tickets.FindByID(ticketID)
	if err != nil {
		return nil, err
	}
	pub := updated.Public()
	return &pub, nil
}

// resolveTicketAction 判断这一步能不能走、走到哪。非法跳转 409，人不对 403。
func resolveTicketAction(t *model.Ticket, actorID int64, role, action string) (to string, clearAssignee bool, err error) {
	isAdmin := role == "admin"
	isAssignee := t.AssigneeID != nil && *t.AssigneeID == actorID
	isCreator := t.CreatorID == actorID

	switch action {
	case model.ActionStart:
		// assigned → in_progress：当前处理人或管理员。
		if t.Status != model.StatusAssigned {
			return "", false, conflict("当前状态不能开始处理")
		}
		if !isAdmin && !isAssignee {
			return "", false, permissionDenied("只有当前处理人或管理员可以开始处理")
		}
		return model.StatusInProgress, false, nil
	case model.ActionResolve:
		// in_progress → resolved：当前处理人或管理员。
		if t.Status != model.StatusInProgress {
			return "", false, conflict("当前状态不能标记已解决")
		}
		if !isAdmin && !isAssignee {
			return "", false, permissionDenied("只有当前处理人或管理员可以标记已解决")
		}
		return model.StatusResolved, false, nil
	case model.ActionClose:
		// resolved → closed：提单人或管理员。
		if t.Status != model.StatusResolved {
			return "", false, conflict("当前状态不能关闭")
		}
		if !isAdmin && !isCreator {
			return "", false, permissionDenied("只有提单人或管理员可以关闭工单")
		}
		return model.StatusClosed, false, nil
	case model.ActionReopen:
		// resolved → open，并清空处理人：提单人或管理员。
		if t.Status != model.StatusResolved {
			return "", false, conflict("当前状态不能重开")
		}
		if !isAdmin && !isCreator {
			return "", false, permissionDenied("只有提单人或管理员可以重开工单")
		}
		return model.StatusOpen, true, nil
	default:
		return "", false, invalidArg("动作必须是 start / resolve / close / reopen")
	}
}

// Comment 对可见工单留言。closed 不能评；看不见当 404。评论本身不写审计。
func (s *TicketService) Comment(actorID, ticketID int64, role, body string) (*model.PublicComment, error) {
	body = strings.TrimSpace(body)
	n := utf8.RuneCountInString(body)
	if n == 0 || n > 2000 {
		return nil, invalidArg("评论必填，最多 2000 个字")
	}

	t, err := s.tickets.GetDetailByViewer(actorID, ticketID, role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	if t.Status == model.StatusClosed {
		return nil, conflict("工单已关闭，不能再评论")
	}

	c, err := s.comments.Insert(ticketID, actorID, body)
	if err != nil {
		return nil, err
	}
	pub := c.Public()
	return &pub, nil
}
