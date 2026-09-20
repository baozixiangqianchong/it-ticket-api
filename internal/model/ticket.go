package model

import "time"

const (
	CategoryHardware = "hardware"
	CategorySoftware = "software"
	CategoryNetwork  = "network"
	CategoryOther    = "other"

	StatusOpen = "open"

	AuditCreate = "create"
)

// Ticket 对应 tickets 表一行。
type Ticket struct {
	ID          int64
	Title       string
	Description string
	Category    string
	Status      string
	CreatorID   int64
	AssigneeID  *int64 // 未指派时为 nil
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PublicTicket 接口返回的工单。创建接口不带评论和审计时间线。
type PublicTicket struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	CreatorID   int64     `json:"creator_id"`
	AssigneeID  *int64    `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t Ticket) Public() PublicTicket {
	return PublicTicket{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Status:      t.Status,
		CreatorID:   t.CreatorID,
		AssigneeID:  t.AssigneeID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// CreateTicketInput 创建工单入参。客户端就算传 status / creator_id 也会被忽略。
type CreateTicketInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
