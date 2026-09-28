package store

import (
	"database/sql"
	"strconv"
	"strings"

	"it-ticket-api/internal/model"
)

const activityUnion = `
SELECT
	a.id,
	'ticket' AS kind,
	a.action,
	a.actor_id,
	actor.display_name AS actor_name,
	a.ticket_id,
	t.title AS ticket_title,
	NULL AS user_id,
	NULL AS user_name,
	a.from_status AS from_value,
	a.to_status AS to_value,
	a.reason,
	a.created_at
FROM audit_logs a
JOIN users actor ON actor.id = a.actor_id
JOIN tickets t ON t.id = a.ticket_id
UNION ALL
SELECT
	aa.id,
	'account' AS kind,
	aa.action,
	aa.actor_id,
	actor.display_name AS actor_name,
	NULL AS ticket_id,
	NULL AS ticket_title,
	aa.user_id,
	target.display_name AS user_name,
	aa.from_value,
	aa.to_value,
	NULL AS reason,
	aa.created_at
FROM account_audits aa
JOIN users actor ON actor.id = aa.actor_id
JOIN users target ON target.id = aa.user_id`

type ActivityStore struct {
	db *sql.DB
}

func NewActivityStore(db *sql.DB) *ActivityStore {
	return &ActivityStore{db: db}
}

func (s *ActivityStore) InsertAccount(tx *sql.Tx, actorID, userID int64, action string, from *string, to string) error {
	_, err := tx.Exec(
		`INSERT INTO account_audits (actor_id, user_id, action, from_value, to_value) VALUES (?, ?, ?, ?, ?)`,
		actorID, userID, action, from, to,
	)
	return err
}

func (s *ActivityStore) Count(f model.ActivityFilter) (int64, error) {
	where, args := activityFilterSQL(f)
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM (`+activityUnion+`) x`+where, args...).Scan(&n)
	return n, err
}

func (s *ActivityStore) List(f model.ActivityFilter) ([]model.Activity, error) {
	where, args := activityFilterSQL(f)
	query := `SELECT id, kind, action, actor_id, actor_name, ticket_id, ticket_title, user_id, user_name, from_value, to_value, reason, created_at FROM (` + activityUnion + `) x` + where + ` ORDER BY created_at DESC, id DESC`
	if f.Limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Activity, 0)
	for rows.Next() {
		var a model.Activity
		var ticketID, userID sql.NullInt64
		var ticketTitle, userName, from, reason sql.NullString
		if err := rows.Scan(
			&a.ID, &a.Kind, &a.Action, &a.ActorID, &a.ActorName,
			&ticketID, &ticketTitle, &userID, &userName, &from, &a.ToValue, &reason, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		if ticketID.Valid {
			a.TicketID = &ticketID.Int64
		}
		if ticketTitle.Valid {
			a.TicketTitle = &ticketTitle.String
		}
		if userID.Valid {
			a.UserID = &userID.Int64
		}
		if userName.Valid {
			a.UserName = &userName.String
		}
		if from.Valid {
			a.FromValue = &from.String
		}
		if reason.Valid {
			a.Reason = &reason.String
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func activityFilterSQL(f model.ActivityFilter) (string, []any) {
	var parts []string
	var args []any
	if kind := strings.TrimSpace(f.Kind); kind != "" {
		parts = append(parts, "kind = ?")
		args = append(args, kind)
	}
	if action := strings.TrimSpace(f.Action); action != "" {
		parts = append(parts, "action = ?")
		args = append(args, action)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		like := "%" + escapeLike(q) + "%"
		raw := strings.TrimPrefix(q, "#")
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 && strconv.FormatInt(id, 10) == raw {
			parts = append(parts, "(ticket_id = ? OR user_id = ? OR actor_name LIKE ? OR IFNULL(ticket_title,'') LIKE ? OR IFNULL(user_name,'') LIKE ?)")
			args = append(args, id, id, like, like, like)
		} else {
			parts = append(parts, "(actor_name LIKE ? OR IFNULL(ticket_title,'') LIKE ? OR IFNULL(user_name,'') LIKE ?)")
			args = append(args, like, like, like)
		}
	}
	if f.From != nil {
		parts = append(parts, "created_at >= ?")
		args = append(args, *f.From)
	}
	if f.To != nil {
		parts = append(parts, "created_at < ?")
		args = append(args, *f.To)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}
