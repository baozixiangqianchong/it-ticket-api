package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

const (
	maxTemplates = 20
	maxReplies   = 30
	maxFields    = 8
)

type CatalogService struct {
	catalog *store.CatalogStore
}

func NewCatalogService(catalog *store.CatalogStore) *CatalogService {
	return &CatalogService{catalog: catalog}
}

func (s *CatalogService) ListTemplates(enabledOnly bool) (*model.TicketTemplateList, error) {
	list, err := s.catalog.ListTemplates(enabledOnly)
	if err != nil {
		return nil, err
	}
	return &model.TicketTemplateList{Items: list}, nil
}

func (s *CatalogService) CreateTemplate(in model.TemplateInput) (*model.TicketTemplate, error) {
	n, err := s.catalog.CountTemplates()
	if err != nil {
		return nil, err
	}
	if n >= maxTemplates {
		return nil, invalidArg("模板最多 20 个")
	}
	t, err := normalizeTemplate(0, in)
	if err != nil {
		return nil, err
	}
	return s.catalog.InsertTemplate(t)
}

func (s *CatalogService) UpdateTemplate(id int64, in model.TemplateInput) (*model.TicketTemplate, error) {
	if _, err := s.mustTemplate(id); err != nil {
		return nil, err
	}
	t, err := normalizeTemplate(id, in)
	if err != nil {
		return nil, err
	}
	return s.catalog.UpdateTemplate(t)
}

func (s *CatalogService) DeleteTemplate(id int64) error {
	if _, err := s.mustTemplate(id); err != nil {
		return err
	}
	return s.catalog.DeleteTemplate(id)
}

func (s *CatalogService) ListReplies(enabledOnly bool) (*model.CannedReplyList, error) {
	list, err := s.catalog.ListReplies(enabledOnly)
	if err != nil {
		return nil, err
	}
	return &model.CannedReplyList{Items: list}, nil
}

func (s *CatalogService) CreateReply(in model.CannedReplyInput) (*model.CannedReply, error) {
	n, err := s.catalog.CountReplies()
	if err != nil {
		return nil, err
	}
	if n >= maxReplies {
		return nil, invalidArg("常用回复最多 30 条")
	}
	r, err := normalizeReply(0, in)
	if err != nil {
		return nil, err
	}
	return s.catalog.InsertReply(r)
}

func (s *CatalogService) UpdateReply(id int64, in model.CannedReplyInput) (*model.CannedReply, error) {
	if _, err := s.mustReply(id); err != nil {
		return nil, err
	}
	r, err := normalizeReply(id, in)
	if err != nil {
		return nil, err
	}
	return s.catalog.UpdateReply(r)
}

func (s *CatalogService) DeleteReply(id int64) error {
	if _, err := s.mustReply(id); err != nil {
		return err
	}
	return s.catalog.DeleteReply(id)
}

func (s *CatalogService) mustTemplate(id int64) (*model.TicketTemplate, error) {
	t, err := s.catalog.GetTemplate(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("模板不存在")
		}
		return nil, err
	}
	return t, nil
}

func (s *CatalogService) mustReply(id int64) (*model.CannedReply, error) {
	r, err := s.catalog.GetReply(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("常用回复不存在")
		}
		return nil, err
	}
	return r, nil
}

func normalizeTemplate(id int64, in model.TemplateInput) (model.TicketTemplate, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > 32 {
		return model.TicketTemplate{}, invalidArg("模板名称必填，最多 32 个字")
	}
	category := strings.TrimSpace(in.Category)
	if category != model.CategoryHardware && category != model.CategorySoftware && category != model.CategoryNetwork && category != model.CategoryOther {
		return model.TicketTemplate{}, invalidArg("分类必须是 hardware / software / network / other")
	}
	titleHint := strings.TrimSpace(in.TitleHint)
	if titleHint == "" || utf8.RuneCountInString(titleHint) > 120 {
		return model.TicketTemplate{}, invalidArg("默认标题必填，最多 120 个字")
	}
	hint := strings.TrimSpace(in.Hint)
	if utf8.RuneCountInString(hint) > 80 {
		return model.TicketTemplate{}, invalidArg("说明最多 80 个字")
	}
	icon := strings.TrimSpace(in.Icon)
	if icon == "" {
		icon = "other"
	}
	if icon != "printer" && icon != "email" && icon != "network" && icon != "other" {
		return model.TicketTemplate{}, invalidArg("图标必须是 printer / email / network / other")
	}
	fields, err := normalizeFields(in.Fields)
	if err != nil {
		return model.TicketTemplate{}, err
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	return model.TicketTemplate{
		ID:        id,
		Name:      name,
		Category:  category,
		TitleHint: titleHint,
		Hint:      hint,
		Icon:      icon,
		SortOrder: in.SortOrder,
		Enabled:   enabled,
		Fields:    fields,
	}, nil
}

func normalizeFields(in []model.TemplateField) ([]model.TemplateField, error) {
	if len(in) == 0 {
		return nil, invalidArg("至少要有一个字段")
	}
	if len(in) > maxFields {
		return nil, invalidArg("每个模板最多 8 个字段")
	}
	out := make([]model.TemplateField, 0, len(in))
	seen := map[string]bool{}
	for i, f := range in {
		label := strings.TrimSpace(f.Label)
		if label == "" || utf8.RuneCountInString(label) > 32 {
			return nil, invalidArg("字段名称必填，最多 32 个字")
		}
		placeholder := strings.TrimSpace(f.Placeholder)
		if utf8.RuneCountInString(placeholder) > 80 {
			return nil, invalidArg("字段提示最多 80 个字")
		}
		key := strings.ToLower(strings.TrimSpace(f.Key))
		if key == "" {
			key = fmt.Sprintf("field_%d", i+1)
		}
		if !validFieldKey(key) {
			return nil, invalidArg("字段 key 只能是小写字母、数字和下划线")
		}
		if seen[key] {
			return nil, invalidArg("字段 key 不能重复")
		}
		seen[key] = true
		out = append(out, model.TemplateField{Key: key, Label: label, Placeholder: placeholder, Required: f.Required})
	}
	return out, nil
}

func validFieldKey(s string) bool {
	if s == "" || len(s) > 32 {
		return false
	}
	for i, r := range s {
		ok := r == '_' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if i == 0 && (r >= '0' && r <= '9') {
			return false
		}
		if !ok {
			return false
		}
	}
	return true
}

func normalizeReply(id int64, in model.CannedReplyInput) (model.CannedReply, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > 80 {
		return model.CannedReply{}, invalidArg("标题必填，最多 80 个字")
	}
	body := strings.TrimSpace(in.Body)
	n := utf8.RuneCountInString(body)
	if n == 0 || n > 2000 {
		return model.CannedReply{}, invalidArg("回复内容必填，最多 2000 个字")
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	return model.CannedReply{ID: id, Title: title, Body: body, SortOrder: in.SortOrder, Enabled: enabled}, nil
}
