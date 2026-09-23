package store

import (
	"database/sql"

	"it-ticket-api/internal/model"
)

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// Insert 写一条审计。fromStatus / reason 为 nil 表示没有。必须和工单写入同一条事务。
func (s *AuditStore) Insert(tx *sql.Tx, ticketID, actorID int64, action string, fromStatus, reason *string, toStatus string) error {
	_, err := tx.Exec(
		`INSERT INTO audit_logs (ticket_id, actor_id, action, from_status, to_status, reason) VALUES (?, ?, ?, ?, ?, ?)`,
		ticketID, actorID, action, fromStatus, toStatus, reason,
	)
	return err
}

// ListByTicketID 按时间正序拉一张单的审计，给详情时间线用。
func (s *AuditStore) ListByTicketID(ticketID int64) ([]model.Audit, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.ticket_id, a.actor_id, a.action, a.from_status, a.to_status, a.reason, a.created_at, u.display_name
		 FROM audit_logs a
		 JOIN users u ON u.id = a.actor_id
		 WHERE a.ticket_id = ?
		 ORDER BY a.created_at ASC, a.id ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Audit, 0)
	for rows.Next() {
		var a model.Audit
		var from, reason sql.NullString
		if err := rows.Scan(
			&a.ID, &a.TicketID, &a.ActorID, &a.Action,
			&from, &a.ToStatus, &reason, &a.CreatedAt, &a.ActorName,
		); err != nil {
			return nil, err
		}
		if from.Valid {
			a.FromStatus = &from.String
		}
		if reason.Valid {
			a.Reason = &reason.String
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
