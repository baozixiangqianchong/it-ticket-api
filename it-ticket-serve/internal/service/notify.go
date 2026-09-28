package service

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode/utf8"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

func ticketSubject(t *model.Ticket) string {
	return clipRunes(fmt.Sprintf("#%d %s", t.ID, t.Title), 120)
}

func withReason(base, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return base
	}
	return base + "：" + reason
}

func roleCN(role string) string {
	switch role {
	case "user":
		return "员工"
	case "agent":
		return "IT"
	case "admin":
		return "管理员"
	default:
		return role
	}
}

func displayName(users *store.UserStore, id int64) string {
	if users == nil || id <= 0 {
		return "同事"
	}
	u, err := users.FindByID(id)
	if err != nil || strings.TrimSpace(u.DisplayName) == "" {
		return "同事"
	}
	return u.DisplayName
}

func clipRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func (s *TicketService) notifyTicket(tx *sql.Tx, userID, actorID int64, t *model.Ticket, typ, body string) error {
	if s.notifs == nil || userID <= 0 || userID == actorID {
		return nil
	}
	id := t.ID
	return s.notifs.Insert(tx, userID, &id, typ, ticketSubject(t), clipRunes(body, 500))
}

func (s *AdminService) notifyAccount(userID, actorID int64, typ, title, body string) {
	if s.notifs == nil || userID <= 0 || userID == actorID {
		return
	}
	if err := s.notifs.Insert(nil, userID, nil, typ, clipRunes(title, 120), clipRunes(body, 500)); err != nil {
		logger.Warn("写账号通知失败", "user_id", userID, "type", typ, "err", err)
	}
}
