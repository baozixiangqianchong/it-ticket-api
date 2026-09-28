package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"

	"it-ticket-api/internal/model"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Create 创建用户
func (s *UserStore) Create(email, passwordHash, displayName, role string) (*model.User, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (email, password_hash, display_name, role) VALUES (?, ?, ?, ?)`,
		email, passwordHash, displayName, role,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.FindByID(id)
}

func (s *UserStore) CreateTx(tx *sql.Tx, email, passwordHash, displayName, role string) (*model.User, error) {
	res, err := tx.Exec(
		`INSERT INTO users (email, password_hash, display_name, role) VALUES (?, ?, ?, ?)`,
		email, passwordHash, displayName, role,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return scanUser(tx.QueryRow(
		`SELECT id, email, password_hash, display_name, role, status, created_at, updated_at FROM users WHERE id = ?`,
		id,
	))
}

func (s *UserStore) ListActiveHandlers() ([]model.User, error) {
	rows, err := s.db.Query(
		`SELECT id, email, password_hash, display_name, role, status, created_at, updated_at
		 FROM users WHERE role IN ('agent', 'admin') AND status = ? ORDER BY role DESC, id ASC`,
		model.UserActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if u.Status == "" {
			u.Status = model.UserActive
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// FindByID 根据用户 ID 查询用户信息
func (s *UserStore) FindByID(id int64) (*model.User, error) {
	return scanUser(s.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, status, created_at, updated_at FROM users WHERE id = ?`,
		id,
	))
}

// FindByEmail 根据用户邮箱查询用户信息
func (s *UserStore) FindByEmail(email string) (*model.User, error) {
	u, err := scanUser(s.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, status, created_at, updated_at FROM users WHERE email = ?`,
		email,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

// scanUser 将数据库查询结果扫描到 model.User 结构体中
func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	if u.Status == "" {
		u.Status = model.UserActive
	}
	return &u, nil
}

var ErrEmailTaken = errors.New("email taken")

// List 分页查用户。role / q 为空则不筛。
func (s *UserStore) List(offset, limit int, role, status, q string) ([]model.User, error) {
	where, args := userFilter(role, status, q)
	args = append(args, limit, offset)
	rows, err := s.db.Query(
		`SELECT id, email, password_hash, display_name, role, status, created_at, updated_at
		 FROM users`+where+` ORDER BY id ASC LIMIT ? OFFSET ?`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if u.Status == "" {
			u.Status = model.UserActive
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *UserStore) Count(role, status, q string) (int64, error) {
	where, args := userFilter(role, status, q)
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`+where, args...).Scan(&n)
	return n, err
}

func userFilter(role, status, q string) (string, []any) {
	var parts []string
	var args []any
	if role != "" {
		parts = append(parts, "role = ?")
		args = append(args, role)
	}
	if status != "" {
		parts = append(parts, "status = ?")
		args = append(args, status)
	}
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + escapeLike(q) + "%"
		parts = append(parts, "(email LIKE ? OR display_name LIKE ?)")
		args = append(args, like, like)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func (s *UserStore) CountByRole(role string) (int64, error) {
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&n)
	return n, err
}

func (s *UserStore) CountActiveByRole(role string) (int64, error) {
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ? AND status = ?`, role, model.UserActive).Scan(&n)
	return n, err
}

func (s *UserStore) UpdateRole(id int64, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, id)
	return err
}

func (s *UserStore) UpdateStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE users SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *UserStore) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

func (s *UserStore) UpdateRoleTx(tx *sql.Tx, id int64, role string) error {
	_, err := tx.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, id)
	return err
}

func (s *UserStore) UpdateStatusTx(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec(`UPDATE users SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *UserStore) UpdateDisplayName(id int64, displayName string) error {
	_, err := s.db.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, displayName, id)
	return err
}

func (s *UserStore) UpdatePasswordHash(id int64, passwordHash string) error {
	_, err := s.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	return err
}
