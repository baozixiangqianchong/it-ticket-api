package service

import (
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

// TicketService 工单业务。状态机、权限、工单与审计同一事务都放这里，不拼 SQL。
type TicketService struct {
	tickets *store.TicketStore // 写/读 tickets 表
	audits  *store.AuditStore  // 写 audit_logs 表
}

func NewTicketService(tickets *store.TicketStore, audits *store.AuditStore) *TicketService {
	return &TicketService{tickets: tickets, audits: audits}
}

// Create 员工（以及 agent / admin 以员工身份）提单。
// 服务端写死 status=open、assignee 为空；客户端就算传这两个字段也没用。
// 工单行和审计行同一事务：只成功一半会回滚，不会出现「有单无审计」。
func (s *TicketService) Create(creatorID int64, in model.CreateTicketInput) (*model.PublicTicket, error) {
	title, description, category, err := validateCreateTicket(in)
	if err != nil {
		return nil, err
	}

	// 事务必须从这里开：store 只执行 SQL，不决定「这两条要不要绑在一起」。
	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	// defer Rollback：后面任何 return（校验后的 DB 错误）都会撤掉未提交的写入。
	// Commit 成功后 Rollback 是空操作，不会把已经提交的再撤掉。
	defer tx.Rollback()

	id, err := s.tickets.Insert(tx, title, description, category, creatorID)
	if err != nil {
		return nil, err
	}
	// from_status 传 nil：创建没有「原状态」。to_status=open，action=create。
	if err := s.audits.Insert(tx, id, creatorID, model.AuditCreate, nil, model.StatusOpen); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 提交后再按 id 读一遍，才能带上库生成的 created_at / updated_at。
	t, err := s.tickets.FindByID(id)
	if err != nil {
		return nil, err
	}
	// Public() 只返回接口该看的字段（和 User.Public 一个思路）。
	pub := t.Public()
	return &pub, nil
}

// validateCreateTicket 整理并校验入参。空白标题、空描述、非法分类都是 400。
func validateCreateTicket(in model.CreateTicketInput) (title, description, category string, err error) {
	title = strings.TrimSpace(in.Title)
	description = strings.TrimSpace(in.Description)
	category = strings.TrimSpace(in.Category)

	// 标题按「字」计数，中文「显示器不亮」是 5，和库里 VARCHAR(120) 对齐。
	n := utf8.RuneCountInString(title)
	if n == 0 || n > 120 {
		return "", "", "", invalidArg("标题必填，最多 120 个字")
	}
	if description == "" {
		return "", "", "", invalidArg("描述必填")
	}
	if !validTicketCategory(category) {
		return "", "", "", invalidArg("分类必须是 hardware / software / network / other")
	}
	return title, description, category, nil
}

// validTicketCategory 只放行文档里的四个枚举。其它字符串不要等 MySQL ENUM 报错再失败。
func validTicketCategory(s string) bool {
	switch s {
	case model.CategoryHardware, model.CategorySoftware, model.CategoryNetwork, model.CategoryOther:
		return true
	default:
		return false
	}
}

// List 同一条列表：按角色缩小可见范围，不给三种人各开一个接口。
func (s *TicketService) List(userID int64, role string) ([]model.PublicTicket, error) {
	return s.tickets.ListByViewer(userID, role)
}

// GetDetail 工单详情：可见范围和列表相同。看不见或没有这张单，都是 404。
func (s *TicketService) GetDetail(userID, id int64, role string) (*model.PublicTicket, error) {
	t, err := s.tickets.GetDetailByViewer(userID, id, role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("工单不存在")
		}
		return nil, err
	}
	pub := t.Public()
	return &pub, nil
}
