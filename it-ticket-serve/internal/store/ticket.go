package store

import (
	"database/sql"
	"strings"

	"it-ticket-api/internal/model"
)

const ticketSelect = `SELECT t.id, t.title, t.description, t.category, t.status,
	t.creator_id, t.assignee_id, t.created_at, t.updated_at, t.closed_at,
	cu.display_name, COALESCE(au.display_name, '')
 FROM tickets t
 JOIN users cu ON cu.id = t.creator_id
 LEFT JOIN users au ON au.id = t.assignee_id`

// TicketStore 只谈 tickets 表的 SQL。开不开事务、状态能不能跳，由 service 决定。
type TicketStore struct {
	db *sql.DB
}

func NewTicketStore(db *sql.DB) *TicketStore {
	return &TicketStore{db: db}
}

func (s *TicketStore) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

func (s *TicketStore) Insert(tx *sql.Tx, title, description, category string, creatorID int64) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO tickets (title, description, category, status, creator_id) VALUES (?, ?, ?, ?, ?)`,
		title, description, category, model.StatusOpen, creatorID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *TicketStore) FindByID(id int64) (*model.Ticket, error) {
	return scanTicket(s.db.QueryRow(ticketSelect+` WHERE t.id = ?`, id))
}

// GetByViewer 按 id 取一张单，同时套上可见范围。看不见或没有都是 sql.ErrNoRows。
func (s *TicketStore) GetByViewer(userID, id int64, role string) (*model.Ticket, error) {
	clause, args := viewerClause(userID, role)
	query := ticketSelect + ` WHERE t.id = ?`
	allArgs := []any{id}
	if clause != "" {
		query += ` AND (` + clause + `)`
		allArgs = append(allArgs, args...)
	}
	return scanTicket(s.db.QueryRow(query, allArgs...))
}

func (s *TicketStore) List(f model.TicketFilter) ([]model.Ticket, error) {
	where, args := filterSQL(f)
	query := ticketSelect + where + ` ORDER BY t.updated_at DESC, t.id DESC`
	if f.Limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Ticket, 0)
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	return list, rows.Err()
}

func (s *TicketStore) Count(f model.TicketFilter) (int64, error) {
	where, args := filterSQL(f)
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tickets t`+where, args...).Scan(&n)
	return n, err
}

// UpdateAssignee 指派 / 改派：写入处理人，状态写成 assigned。
func (s *TicketStore) UpdateAssignee(tx *sql.Tx, ticketID, assigneeID int64) error {
	_, err := tx.Exec(
		`UPDATE tickets SET assignee_id = ?, status = ?, closed_at = NULL WHERE id = ?`,
		assigneeID, model.StatusAssigned, ticketID,
	)
	return err
}

// ClaimOpen 只有待派且无人领取时才成功。影响 0 行表示被人抢先领了。
func (s *TicketStore) ClaimOpen(tx *sql.Tx, ticketID, assigneeID int64) (bool, error) {
	res, err := tx.Exec(
		`UPDATE tickets SET assignee_id = ?, status = ? WHERE id = ? AND status = ? AND assignee_id IS NULL`,
		assigneeID, model.StatusAssigned, ticketID, model.StatusOpen,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

// ApplyTransition 按状态机写下一步。各标志由 service 算好，这里只拼 UPDATE。
func (s *TicketStore) ApplyTransition(tx *sql.Tx, ticketID int64, u TicketUpdate) error {
	sets := []string{"status = ?"}
	args := []any{u.Status}
	if u.ClearAssignee {
		sets = append(sets, "assignee_id = NULL")
	}
	if u.SetClosedNow {
		sets = append(sets, "closed_at = CURRENT_TIMESTAMP")
	}
	if u.ClearClosedAt {
		sets = append(sets, "closed_at = NULL")
	}
	args = append(args, ticketID)
	_, err := tx.Exec(`UPDATE tickets SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

// TicketUpdate 一次状态变更要写哪些列。
type TicketUpdate struct {
	Status        string
	ClearAssignee bool
	SetClosedNow  bool
	ClearClosedAt bool
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTicket(sc rowScanner) (*model.Ticket, error) {
	var t model.Ticket
	var assignee sql.NullInt64
	var closed sql.NullTime
	if err := sc.Scan(
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Status,
		&t.CreatorID, &assignee, &t.CreatedAt, &t.UpdatedAt, &closed,
		&t.CreatorName, &t.AssigneeName,
	); err != nil {
		return nil, err
	}
	if assignee.Valid {
		t.AssigneeID = &assignee.Int64
	}
	if closed.Valid {
		t.ClosedAt = &closed.Time
	}
	return &t, nil
}

// viewerClause V2：员工只看自己的；IT 看自己提的、派给自己的、以及待派池；管理员不限制。
func viewerClause(userID int64, role string) (string, []any) {
	switch role {
	case "admin":
		return "", nil
	case "agent":
		return "t.creator_id = ? OR t.assignee_id = ? OR (t.status = ? AND t.assignee_id IS NULL)",
			[]any{userID, userID, model.StatusOpen}
	default:
		return "t.creator_id = ?", []any{userID}
	}
}

func filterSQL(f model.TicketFilter) (string, []any) {
	var parts []string
	var args []any

	switch f.Scope {
	case model.ScopeCreated:
		parts = append(parts, "t.creator_id = ?")
		args = append(args, f.ViewerID)
	case model.ScopeAssigned:
		parts = append(parts, "t.assignee_id = ? AND t.status <> ?")
		args = append(args, f.ViewerID, model.StatusClosed)
	case model.ScopePool:
		parts = append(parts, "t.status = ? AND t.assignee_id IS NULL")
		args = append(args, model.StatusOpen)
	default:
		if clause, a := viewerClause(f.ViewerID, f.ViewerRole); clause != "" {
			parts = append(parts, "("+clause+")")
			args = append(args, a...)
		}
	}

	if f.Status != "" {
		parts = append(parts, "t.status = ?")
		args = append(args, f.Status)
	}
	if f.Category != "" {
		parts = append(parts, "t.category = ?")
		args = append(args, f.Category)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		parts = append(parts, "t.title LIKE ?")
		args = append(args, "%"+escapeLike(q)+"%")
	}
	if f.AssigneeID != nil {
		parts = append(parts, "t.assignee_id = ?")
		args = append(args, *f.AssigneeID)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

// escapeLike 去掉通配符，避免用户输入 % 把筛选变成全表。
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}
