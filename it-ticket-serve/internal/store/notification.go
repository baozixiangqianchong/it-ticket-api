package store

import (
	"database/sql"

	"it-ticket-api/internal/model"
)

type NotificationStore struct {
	db *sql.DB
}

func NewNotificationStore(db *sql.DB) *NotificationStore {
	return &NotificationStore{db: db}
}

func (s *NotificationStore) Insert(tx *sql.Tx, userID int64, ticketID *int64, typ, title, body string) error {
	if tx != nil {
		_, err := tx.Exec(
			`INSERT INTO notifications (user_id, ticket_id, type, title, body) VALUES (?, ?, ?, ?, ?)`,
			userID, ticketID, typ, title, body,
		)
		return err
	}
	_, err := s.db.Exec(
		`INSERT INTO notifications (user_id, ticket_id, type, title, body) VALUES (?, ?, ?, ?, ?)`,
		userID, ticketID, typ, title, body,
	)
	return err
}

func (s *NotificationStore) List(userID int64, unreadOnly bool, limit, offset int) ([]model.Notification, error) {
	query := `SELECT id, user_id, ticket_id, type, title, COALESCE(body, ''), read_at, created_at
		FROM notifications WHERE user_id = ?`
	args := []any{userID}
	if unreadOnly {
		query += ` AND read_at IS NULL`
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *n)
	}
	return list, rows.Err()
}

func (s *NotificationStore) Count(userID int64, unreadOnly bool) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ?`
	args := []any{userID}
	if unreadOnly {
		query += ` AND read_at IS NULL`
	}
	var n int64
	err := s.db.QueryRow(query, args...).Scan(&n)
	return n, err
}

func (s *NotificationStore) UnreadCount(userID int64) (int64, error) {
	return s.Count(userID, true)
}

func (s *NotificationStore) MarkRead(userID, id int64) (bool, error) {
	res, err := s.db.Exec(
		`UPDATE notifications SET read_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ? AND read_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (s *NotificationStore) MarkAllRead(userID int64) error {
	_, err := s.db.Exec(
		`UPDATE notifications SET read_at = CURRENT_TIMESTAMP WHERE user_id = ? AND read_at IS NULL`,
		userID,
	)
	return err
}

func (s *NotificationStore) Owned(userID, id int64) (bool, error) {
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE id = ? AND user_id = ?`, id, userID).Scan(&n)
	return n == 1, err
}

func scanNotification(sc rowScanner) (*model.Notification, error) {
	var n model.Notification
	var ticketID sql.NullInt64
	var readAt sql.NullTime
	if err := sc.Scan(&n.ID, &n.UserID, &ticketID, &n.Type, &n.Title, &n.Body, &readAt, &n.CreatedAt); err != nil {
		return nil, err
	}
	if ticketID.Valid {
		n.TicketID = &ticketID.Int64
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	return &n, nil
}
