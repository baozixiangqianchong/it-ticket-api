package store

import (
	"database/sql"
	"errors"

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

// FindByID 根据用户 ID 查询用户信息
func (s *UserStore) FindByID(id int64) (*model.User, error) {
	return scanUser(s.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, created_at, updated_at FROM users WHERE id = ?`,
		id,
	))
}

// FindByEmail 根据用户邮箱查询用户信息
func (s *UserStore) FindByEmail(email string) (*model.User, error) {
	u, err := scanUser(s.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, created_at, updated_at FROM users WHERE email = ?`,
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
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

var ErrEmailTaken = errors.New("email taken")
