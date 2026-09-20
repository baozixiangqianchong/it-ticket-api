package store

import (
	"database/sql"
)

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// Insert 写一条审计。fromStatus == nil 表示创建这种「没有原状态」的动作。
// 必须传入和工单写入同一个 *sql.Tx，不要单独提交。
func (s *AuditStore) Insert(tx *sql.Tx, ticketID, actorID int64, action string, fromStatus *string, toStatus string) error {
	_, err := tx.Exec(
		`INSERT INTO audit_logs (ticket_id, actor_id, action, from_status, to_status) VALUES (?, ?, ?, ?, ?)`,
		ticketID, actorID, action, fromStatus, toStatus,
	)
	return err
}
