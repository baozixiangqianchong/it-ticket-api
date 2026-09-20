package store

import (
	"database/sql"

	"it-ticket-api/internal/model"
)

// TicketStore 只谈 tickets 表的 SQL。开不开事务、状态能不能跳，由 service 决定。
type TicketStore struct {
	db *sql.DB // 连接池，FindByID 走它；Insert 必须走传入的 tx，才能和审计同一事务
}

func NewTicketStore(db *sql.DB) *TicketStore {
	return &TicketStore{db: db}
}

// Begin 开事务。工单 INSERT 和审计 INSERT 必须一起提交，由 service 调用本方法。
func (s *TicketStore) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

// Insert 写入一张待派单。status 固定 open；不写 assignee_id，库里就是 NULL。
// 必须用 tx.Exec，不能用 s.db.Exec，否则审计失败时这条工单已经独立提交，滚不回去。
func (s *TicketStore) Insert(tx *sql.Tx, title, description, category string, creatorID int64) (int64, error) {
	// ? 是占位符，按后面参数顺序填，避免手拼 SQL 被注入。
	res, err := tx.Exec(
		`INSERT INTO tickets (title, description, category, status, creator_id) VALUES (?, ?, ?, ?, ?)`,
		title, description, category, model.StatusOpen, creatorID,
	)
	if err != nil {
		return 0, err
	}
	// 自增主键，给审计的 ticket_id 和后面 FindByID 用。
	return res.LastInsertId()
}

// FindByID 按主键读一行。创建提交后用它带回 created_at / updated_at。
// 这里走连接池 s.db，不走 tx：调用方已经 Commit，事务结束了。
func (s *TicketStore) FindByID(id int64) (*model.Ticket, error) {
	row := s.db.QueryRow(
		`SELECT id, title, description, category, status, creator_id, assignee_id, created_at, updated_at
		 FROM tickets WHERE id = ?`,
		id,
	)
	return scanTicket(row)
}

// scanTicket 把一行扫进结构体。assignee_id 可空，不能直接扫进 *int64，要先接 NullInt64。
func scanTicket(row *sql.Row) (*model.Ticket, error) {
	var t model.Ticket
	var assignee sql.NullInt64
	if err := row.Scan(
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Status,
		&t.CreatorID, &assignee, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	// Valid=false 表示 SQL NULL（未指派），AssigneeID 保持 nil，JSON 会是 null。
	if assignee.Valid {
		t.AssigneeID = &assignee.Int64
	}
	return &t, nil
}
