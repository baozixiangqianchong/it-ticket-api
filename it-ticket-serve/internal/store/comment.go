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
		`SELECT c.id, c.ticket_id, c.author_id, c.body, c.created_at, u.display_name
		 FROM ticket_comments c
		 JOIN users u ON u.id = c.author_id
		 WHERE c.id = ?`,
		id,
	)
	return scanComment(row)
}

func (s *CommentStore) ListByTicketID(ticketID int64) ([]model.Comment, error) {
	rows, err := s.db.Query(
		`SELECT c.id, c.ticket_id, c.author_id, c.body, c.created_at, u.display_name
		 FROM ticket_comments c
		 JOIN users u ON u.id = c.author_id
		 WHERE c.ticket_id = ?
		 ORDER BY c.created_at ASC, c.id ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Comment, 0)
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *c)
	}
	return list, rows.Err()
}

func scanComment(sc rowScanner) (*model.Comment, error) {
	var c model.Comment
	if err := sc.Scan(&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.AuthorName); err != nil {
		return nil, err
	}
	return &c, nil
}
