package store

import (
	"database/sql"

	"it-ticket-api/internal/model"
)

type CommentStore struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{db: db}
}

// Insert 追加一条评论。作者和工单 id 由 service 传入，不从客户端信。
func (s *CommentStore) Insert(ticketID, authorID int64, body string) (*model.Comment, error) {
	res, err := s.db.Exec(
		`INSERT INTO ticket_comments (ticket_id, author_id, body) VALUES (?, ?, ?)`,
		ticketID, authorID, body,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.FindByID(id)
}

func (s *CommentStore) FindByID(id int64) (*model.Comment, error) {
	row := s.db.QueryRow(
		`SELECT id, ticket_id, author_id, body, created_at FROM ticket_comments WHERE id = ?`,
		id,
	)
	var c model.Comment
	if err := row.Scan(&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}
