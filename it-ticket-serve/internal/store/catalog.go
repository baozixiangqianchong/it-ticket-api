package store

import (
	"database/sql"
	"encoding/json"

	"it-ticket-api/internal/model"
)

type CatalogStore struct {
	db *sql.DB
}

func NewCatalogStore(db *sql.DB) *CatalogStore {
	return &CatalogStore{db: db}
}

func (s *CatalogStore) ListTemplates(enabledOnly bool) ([]model.TicketTemplate, error) {
	query := `SELECT id, name, category, title_hint, hint, icon, sort_order, enabled, fields
		 FROM ticket_templates`
	args := []any{}
	if enabledOnly {
		query += ` WHERE enabled = 1`
	}
	query += ` ORDER BY sort_order ASC, id ASC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.TicketTemplate, 0)
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	return list, rows.Err()
}

func (s *CatalogStore) GetTemplate(id int64) (*model.TicketTemplate, error) {
	return scanTemplate(s.db.QueryRow(
		`SELECT id, name, category, title_hint, hint, icon, sort_order, enabled, fields
		 FROM ticket_templates WHERE id = ?`, id,
	))
}

func (s *CatalogStore) InsertTemplate(t model.TicketTemplate) (*model.TicketTemplate, error) {
	raw, err := json.Marshal(t.Fields)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(
		`INSERT INTO ticket_templates (name, category, title_hint, hint, icon, sort_order, enabled, fields)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Category, t.TitleHint, t.Hint, t.Icon, t.SortOrder, boolToInt(t.Enabled), raw,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetTemplate(id)
}

func (s *CatalogStore) UpdateTemplate(t model.TicketTemplate) (*model.TicketTemplate, error) {
	raw, err := json.Marshal(t.Fields)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(
		`UPDATE ticket_templates SET name = ?, category = ?, title_hint = ?, hint = ?, icon = ?, sort_order = ?, enabled = ?, fields = ?
		 WHERE id = ?`,
		t.Name, t.Category, t.TitleHint, t.Hint, t.Icon, t.SortOrder, boolToInt(t.Enabled), raw, t.ID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetTemplate(t.ID)
}

func (s *CatalogStore) DeleteTemplate(id int64) error {
	_, err := s.db.Exec(`DELETE FROM ticket_templates WHERE id = ?`, id)
	return err
}

func (s *CatalogStore) CountTemplates() (int64, error) {
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ticket_templates`).Scan(&n)
	return n, err
}

func (s *CatalogStore) ListReplies(enabledOnly bool) ([]model.CannedReply, error) {
	query := `SELECT id, title, body, sort_order, enabled FROM canned_replies`
	if enabledOnly {
		query += ` WHERE enabled = 1`
	}
	query += ` ORDER BY sort_order ASC, id ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.CannedReply, 0)
	for rows.Next() {
		item, err := scanReply(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	return list, rows.Err()
}

func (s *CatalogStore) GetReply(id int64) (*model.CannedReply, error) {
	return scanReply(s.db.QueryRow(`SELECT id, title, body, sort_order, enabled FROM canned_replies WHERE id = ?`, id))
}

func (s *CatalogStore) InsertReply(r model.CannedReply) (*model.CannedReply, error) {
	res, err := s.db.Exec(
		`INSERT INTO canned_replies (title, body, sort_order, enabled) VALUES (?, ?, ?, ?)`,
		r.Title, r.Body, r.SortOrder, boolToInt(r.Enabled),
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetReply(id)
}

func (s *CatalogStore) UpdateReply(r model.CannedReply) (*model.CannedReply, error) {
	_, err := s.db.Exec(
		`UPDATE canned_replies SET title = ?, body = ?, sort_order = ?, enabled = ? WHERE id = ?`,
		r.Title, r.Body, r.SortOrder, boolToInt(r.Enabled), r.ID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetReply(r.ID)
}

func (s *CatalogStore) DeleteReply(id int64) error {
	_, err := s.db.Exec(`DELETE FROM canned_replies WHERE id = ?`, id)
	return err
}

func (s *CatalogStore) CountReplies() (int64, error) {
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM canned_replies`).Scan(&n)
	return n, err
}

func scanTemplate(sc interface{ Scan(dest ...any) error }) (*model.TicketTemplate, error) {
	var t model.TicketTemplate
	var enabled int
	var raw []byte
	if err := sc.Scan(&t.ID, &t.Name, &t.Category, &t.TitleHint, &t.Hint, &t.Icon, &t.SortOrder, &enabled, &raw); err != nil {
		return nil, err
	}
	t.Enabled = enabled == 1
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &t.Fields); err != nil {
			return nil, err
		}
	}
	if t.Fields == nil {
		t.Fields = []model.TemplateField{}
	}
	return &t, nil
}

func scanReply(sc interface{ Scan(dest ...any) error }) (*model.CannedReply, error) {
	var r model.CannedReply
	var enabled int
	if err := sc.Scan(&r.ID, &r.Title, &r.Body, &r.SortOrder, &enabled); err != nil {
		return nil, err
	}
	r.Enabled = enabled == 1
	return &r, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
