package store

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"it-ticket-api/internal/model"
)

const ticketSelect = `SELECT t.id, t.title, t.description, t.category, t.priority, t.status,
	t.creator_id, t.assignee_id, t.created_at, t.updated_at, t.closed_at,
	cu.display_name, COALESCE(au.display_name, '')
 FROM tickets t
 JOIN users cu ON cu.id = t.creator_id
 LEFT JOIN users au ON au.id = t.assignee_id`

const ticketListSelect = `SELECT t.id, t.title, t.description, t.category, t.priority, t.status,
	t.creator_id, t.assignee_id, t.created_at, t.updated_at, t.closed_at,
	cu.display_name, COALESCE(au.display_name, ''),
	lc.body, lc.created_at, lc.author_id, COALESCE(lcu.display_name, '')
 FROM tickets t
 JOIN users cu ON cu.id = t.creator_id
 LEFT JOIN users au ON au.id = t.assignee_id
 LEFT JOIN ticket_comments lc ON lc.id = (
	SELECT MAX(id) FROM ticket_comments WHERE ticket_id = t.id
 )
 LEFT JOIN users lcu ON lcu.id = lc.author_id`

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

func (s *TicketStore) Insert(tx *sql.Tx, title, description, category, priority string, creatorID int64) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO tickets (title, description, category, priority, status, creator_id) VALUES (?, ?, ?, ?, ?, ?)`,
		title, description, category, priority, model.StatusOpen, creatorID,
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
	query := ticketListSelect + where + ` ORDER BY FIELD(t.priority, 'p1', 'p2', 'p3'), t.updated_at DESC, t.id DESC`
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
		t, err := scanTicketList(rows)
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

// UpdateAssignee 指派 / 转派：写入处理人，状态由调用方决定（进行中或保持等用户）。
func (s *TicketStore) UpdateAssignee(tx *sql.Tx, ticketID, assigneeID int64, status string) error {
	_, err := tx.Exec(
		`UPDATE tickets SET assignee_id = ?, status = ?, closed_at = NULL WHERE id = ?`,
		assigneeID, status, ticketID,
	)
	return err
}

// ClaimOpen 只有待派且无人领取时才成功。影响 0 行表示被人抢先领了。
func (s *TicketStore) ClaimOpen(tx *sql.Tx, ticketID, assigneeID int64) (bool, error) {
	res, err := tx.Exec(
		`UPDATE tickets SET assignee_id = ?, status = ? WHERE id = ? AND status = ? AND assignee_id IS NULL`,
		assigneeID, model.StatusInProgress, ticketID, model.StatusOpen,
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

func (s *TicketStore) ListOpenAssignedTx(tx *sql.Tx, assigneeID int64) ([]model.Ticket, error) {
	rows, err := tx.Query(ticketSelect+` WHERE t.assignee_id = ? AND t.status <> ? ORDER BY t.id ASC`, assigneeID, model.StatusClosed)
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

func (s *TicketStore) CountOpenByAssignees(ids []int64) (map[int64]int64, error) {
	out := make(map[int64]int64, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+1)
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, model.StatusClosed)
	rows, err := s.db.Query(
		`SELECT assignee_id, COUNT(*) FROM tickets WHERE assignee_id IN (`+strings.Join(placeholders, ",")+`) AND status <> ? GROUP BY assignee_id`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, n int64
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
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
	return scanTicketRow(sc, false)
}

func scanTicketList(sc rowScanner) (*model.Ticket, error) {
	return scanTicketRow(sc, true)
}

func scanTicketRow(sc rowScanner, withComment bool) (*model.Ticket, error) {
	var t model.Ticket
	var assignee sql.NullInt64
	var closed sql.NullTime
	dest := []any{
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
		&t.CreatorID, &assignee, &t.CreatedAt, &t.UpdatedAt, &closed,
		&t.CreatorName, &t.AssigneeName,
	}
	var body sql.NullString
	var at sql.NullTime
	var authorID sql.NullInt64
	var author sql.NullString
	if withComment {
		dest = append(dest, &body, &at, &authorID, &author)
	}
	if err := sc.Scan(dest...); err != nil {
		return nil, err
	}
	if assignee.Valid {
		t.AssigneeID = &assignee.Int64
	}
	if closed.Valid {
		t.ClosedAt = &closed.Time
	}
	if withComment && body.Valid && at.Valid && strings.TrimSpace(body.String) != "" {
		t.LastCommentBody = body.String
		t.LastCommentAt = &at.Time
		if authorID.Valid {
			t.LastCommentAuthorID = &authorID.Int64
		}
		if author.Valid {
			t.LastCommentAuthor = author.String
		}
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
		parts = append(parts, "t.assignee_id = ? AND t.status IN (?, ?)")
		args = append(args, f.ViewerID, model.StatusAssigned, model.StatusInProgress)
	case model.ScopeWaiting:
		parts = append(parts, "t.assignee_id = ? AND t.status IN (?, ?)")
		args = append(args, f.ViewerID, model.StatusPending, model.StatusResolved)
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
	if f.Priority != "" {
		parts = append(parts, "t.priority = ?")
		args = append(args, f.Priority)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		like := "%" + escapeLike(q) + "%"
		commentMatch := "EXISTS (SELECT 1 FROM ticket_comments c WHERE c.ticket_id = t.id AND c.body LIKE ?)"
		raw := strings.TrimPrefix(q, "#")
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 && strconv.FormatInt(id, 10) == raw {
			parts = append(parts, "(t.id = ? OR t.title LIKE ? OR t.description LIKE ? OR "+commentMatch+")")
			args = append(args, id, like, like, like)
		} else {
			parts = append(parts, "(t.title LIKE ? OR t.description LIKE ? OR "+commentMatch+")")
			args = append(args, like, like, like)
		}
	}
	if f.AssigneeID != nil {
		parts = append(parts, "t.assignee_id = ?")
		args = append(args, *f.AssigneeID)
	}
	if f.From != nil || f.To != nil {
		col := ticketTimeColumn(f.TimeField)
		if f.From != nil {
			parts = append(parts, col+" >= ?")
			args = append(args, *f.From)
		}
		if f.To != nil {
			parts = append(parts, col+" < ?")
			args = append(args, *f.To)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

// escapeLike 去掉通配符，避免用户输入 % 把筛选变成全表。
func (s *TicketStore) CountStaleOpen(before time.Time) (int64, error) {
	var n int64
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM tickets WHERE status = ? AND assignee_id IS NULL AND created_at < ?`,
		model.StatusOpen, before,
	).Scan(&n)
	return n, err
}

func (s *TicketStore) CountStaleResolved(before time.Time) (int64, error) {
	var n int64
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM tickets WHERE status = ? AND updated_at < ?`,
		model.StatusResolved, before,
	).Scan(&n)
	return n, err
}

func (s *TicketStore) ListStaleResolved(before time.Time) ([]model.Ticket, error) {
	rows, err := s.db.Query(ticketSelect+` WHERE t.status = ? AND t.updated_at < ? ORDER BY t.updated_at ASC, t.id ASC`, model.StatusResolved, before)
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

func (s *TicketStore) AgentLoads() ([]model.AgentLoad, error) {
	rows, err := s.db.Query(
		`SELECT u.id, u.email, u.display_name,
			COALESCE(SUM(t.status IN ('assigned','in_progress')), 0),
			COALESCE(SUM(t.status IN ('pending','resolved')), 0)
		 FROM users u
		 LEFT JOIN tickets t ON t.assignee_id = u.id
		 WHERE u.role IN ('agent', 'admin') AND u.status = ?
		 GROUP BY u.id, u.email, u.display_name
		 ORDER BY u.id ASC`,
		model.UserActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.AgentLoad, 0)
	for rows.Next() {
		var a model.AgentLoad
		if err := rows.Scan(&a.ID, &a.Email, &a.DisplayName, &a.Active, &a.Waiting); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *TicketStore) UpdateFields(tx *sql.Tx, ticketID int64, title, description, category, priority string) error {
	_, err := tx.Exec(
		`UPDATE tickets SET title = ?, description = ?, category = ?, priority = ? WHERE id = ?`,
		title, description, category, priority, ticketID,
	)
	return err
}

func ticketTimeColumn(field string) string {
	switch field {
	case "updated_at":
		return "t.updated_at"
	case "closed_at":
		return "t.closed_at"
	default:
		return "t.created_at"
	}
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}
