package service

import (
	"strings"
	"unicode/utf8"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/model"
)

func (s *TicketService) Edit(actorID, ticketID int64, role string, in model.EditTicketInput) (*model.PublicTicket, error) {
	t, err := s.visibleTicket(actorID, ticketID, role)
	if err != nil {
		return nil, err
	}

	title, description, category, priority, summary, err := prepareTicketEdit(t, actorID, role, in)
	if err != nil {
		return nil, err
	}
	if summary == "" {
		return s.publicByID(ticketID, actorID, role)
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.tickets.UpdateFields(tx, ticketID, title, description, category, priority); err != nil {
		return nil, err
	}
	from := t.Status
	if err := s.audits.Insert(tx, ticketID, actorID, model.AuditEdit, &from, reasonPtr(summary), t.Status); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("工单已修改", "ticket_id", ticketID, "uid", actorID, "change", summary)
	return s.publicByID(ticketID, actorID, role)
}

func prepareTicketEdit(t *model.Ticket, actorID int64, role string, in model.EditTicketInput) (title, description, category, priority, summary string, err error) {
	if t.Status == model.StatusClosed {
		return "", "", "", "", "", ticketClosed("工单已关闭，不能再改")
	}

	isAdmin := role == "admin"
	isCreator := t.CreatorID == actorID
	isAssignee := t.AssigneeID != nil && *t.AssigneeID == actorID
	fullEdit := t.Status == model.StatusOpen && (isCreator || isAdmin)
	priorityOnly := !fullEdit && (isAdmin || isAssignee)
	if !fullEdit && !priorityOnly {
		return "", "", "", "", "", permissionDenied("当前不能修改这张工单")
	}

	title = strings.TrimSpace(in.Title)
	description = strings.TrimSpace(in.Description)
	category = strings.TrimSpace(in.Category)
	priority = strings.TrimSpace(in.Priority)
	if priority == "" {
		priority = t.Priority
	}

	if !fullEdit {
		title = t.Title
		description = t.Description
		category = t.Category
	} else {
		n := utf8.RuneCountInString(title)
		if n == 0 || n > 120 {
			return "", "", "", "", "", invalidArg("标题必填，最多 120 个字")
		}
		if description == "" {
			return "", "", "", "", "", invalidArg("描述必填")
		}
		if !validTicketCategory(category) {
			return "", "", "", "", "", invalidArg("分类必须是 hardware / software / network / other")
		}
	}
	if !validTicketPriority(priority) {
		return "", "", "", "", "", invalidArg("优先级必须是 p1 / p2 / p3")
	}
	if priority == model.PriorityP1 && !isAdmin && t.Priority != model.PriorityP1 {
		return "", "", "", "", "", permissionDenied("紧急工单只能由管理员标记")
	}

	var parts []string
	if title != t.Title {
		parts = append(parts, "标题")
	}
	if description != t.Description {
		parts = append(parts, "描述")
	}
	if category != t.Category {
		parts = append(parts, "分类")
	}
	if priority != t.Priority {
		parts = append(parts, "优先级")
	}
	if len(parts) == 0 {
		return title, description, category, priority, "", nil
	}
	return title, description, category, priority, "修改了" + strings.Join(parts, "、"), nil
}

func denyP1(role, priority string) error {
	if priority == model.PriorityP1 && role != "admin" {
		return permissionDenied("紧急工单只能由管理员标记")
	}
	return nil
}
