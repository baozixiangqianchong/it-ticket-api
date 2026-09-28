package store

import (
	"database/sql"
	"time"

	"it-ticket-api/internal/model"
)

type InviteStore struct {
	db *sql.DB
}

func NewInviteStore(db *sql.DB) *InviteStore {
	return &InviteStore{db: db}
}

func (s *InviteStore) Insert(createdBy int64, code string, expiresAt time.Time) (*model.InviteCode, error) {
	res, err := s.db.Exec(
		`INSERT INTO invite_codes (code, created_by, expires_at) VALUES (?, ?, ?)`,
		code, createdBy, expiresAt,
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

func (s *InviteStore) FindByID(id int64) (*model.InviteCode, error) {
	return scanInvite(s.db.QueryRow(inviteSelect+` WHERE i.id = ?`, id))
}

func (s *InviteStore) FindByCodeForUpdate(tx *sql.Tx, code string) (*model.InviteCode, error) {
	row := tx.QueryRow(inviteSelect+` WHERE i.code = ? FOR UPDATE`, code)
	inv, err := scanInvite(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

func (s *InviteStore) ConsumeTx(tx *sql.Tx, id, usedBy int64) error {
	_, err := tx.Exec(
		`UPDATE invite_codes SET used_at = CURRENT_TIMESTAMP, used_by = ? WHERE id = ? AND used_at IS NULL`,
		usedBy, id,
	)
	return err
}

func (s *InviteStore) List(limit int) ([]model.InviteCode, error) {
	if limit < 1 {
		limit = 30
	}
	rows, err := s.db.Query(inviteSelect+` ORDER BY i.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.InviteCode, 0)
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *inv)
	}
	return list, rows.Err()
}

const inviteSelect = `SELECT i.id, i.code, i.created_by, creator.display_name, i.expires_at,
	i.used_at, i.used_by, used.display_name, i.created_at
 FROM invite_codes i
 JOIN users creator ON creator.id = i.created_by
 LEFT JOIN users used ON used.id = i.used_by`

func scanInvite(sc interface{ Scan(dest ...any) error }) (*model.InviteCode, error) {
	var inv model.InviteCode
	var usedAt sql.NullTime
	var usedBy sql.NullInt64
	var usedName sql.NullString
	if err := sc.Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.CreatorName, &inv.ExpiresAt, &usedAt, &usedBy, &usedName, &inv.CreatedAt); err != nil {
		return nil, err
	}
	if usedAt.Valid {
		inv.UsedAt = &usedAt.Time
	}
	if usedBy.Valid {
		inv.UsedBy = &usedBy.Int64
	}
	if usedName.Valid {
		inv.UsedName = &usedName.String
	}
	return &inv, nil
}
