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

// ListByViewer 按「当前这个人能看什么」查列表。role 是 users 表上的角色，不是 tickets 的列。
// 不要写成 WHERE role = ?：工单行上没有角色字段。
func (s *TicketStore) ListByViewer(userID int64, role string) ([]model.PublicTicket, error) {
	// 三种角色共用同一条 SELECT，差别只在 WHERE。
	query := `SELECT id, title, description, category, status, creator_id, assignee_id, created_at, updated_at FROM tickets`
	var args []any
	switch role {
	case "admin":
		// 管理员：不按人过滤，看全部。
	case "agent":
		// IT：自己提的，或派给自己的。
		query += ` WHERE creator_id = ? OR assignee_id = ?`
		args = []any{userID, userID}
	default:
		// 员工（以及未知角色按员工收）：只能看自己提的。
		query += ` WHERE creator_id = ?`
		args = []any{userID}
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return scanTickets(rows)
}

// scanTickets 把多行扫进切片。assignee_id 可空，和单行 scanTicket 一样先接 NullInt64。
func scanTickets(rows *sql.Rows) ([]model.PublicTicket, error) {
	defer rows.Close()
	tickets := make([]model.PublicTicket, 0)
	for rows.Next() {
		var t model.Ticket
		var assignee sql.NullInt64
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Category, &t.Status,
			&t.CreatorID, &assignee, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if assignee.Valid {
			t.AssigneeID = &assignee.Int64
		}
		tickets = append(tickets, t.Public())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

// GetDetailByViewer 按 id 取一张单，同时套上和列表相同的可见范围。
// 看不见或 id 不存在都会变成 sql.ErrNoRows，service 统一翻成 404，避免泄露「单子在不在」。
// 第一条 WHERE 已经占住了，后面只能 AND，不能再写 WHERE。
func (s *TicketStore) GetDetailByViewer(userID, id int64, role string) (*model.Ticket, error) {
	query := `SELECT id, title, description, category, status, creator_id, assignee_id, created_at, updated_at
		 FROM tickets WHERE id = ?`
	args := []any{id}
	switch role {
	case "admin":
		// 管理员：只要 id 对就能看。
	case "agent":
		query += ` AND (creator_id = ? OR assignee_id = ?)`
		args = append(args, userID, userID)
	default:
		query += ` AND creator_id = ?`
		args = append(args, userID)
	}
	return scanTicket(s.db.QueryRow(query, args...))
}

// UpdateAssignee 把处理人改成指定的 agent，状态写成 assigned。必须走同一条事务。
func (s *TicketStore) UpdateAssignee(tx *sql.Tx, ticketID, assigneeID int64) error {
	_, err := tx.Exec(
		`UPDATE tickets SET assignee_id = ?, status = ? WHERE id = ?`,
		assigneeID, model.StatusAssigned, ticketID,
	)
	return err
}

// UpdateStatus 按状态机写下一步。clearAssignee 为 true 时清空处理人（重开）。
func (s *TicketStore) UpdateStatus(tx *sql.Tx, ticketID int64, status string, clearAssignee bool) error {
	if clearAssignee {
		_, err := tx.Exec(`UPDATE tickets SET status = ?, assignee_id = NULL WHERE id = ?`, status, ticketID)
		return err
	}
	_, err := tx.Exec(`UPDATE tickets SET status = ? WHERE id = ?`, status, ticketID)
	return err
}
