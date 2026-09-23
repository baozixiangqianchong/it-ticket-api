package model

import "time"

const (
	CategoryHardware = "hardware"
	CategorySoftware = "software"
	CategoryNetwork  = "network"
	CategoryOther    = "other"

	StatusOpen       = "open"
	StatusAssigned   = "assigned"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"

	AuditCreate  = "create"
	AuditAssign  = "assign"
	AuditClaim   = "claim"
	AuditStart   = "start"
	AuditResolve = "resolve"
	AuditClose   = "close"
	AuditReopen  = "reopen"
	AuditCancel  = "cancel"

	ActionStart   = "start"
	ActionResolve = "resolve"
	ActionClose   = "close"
	ActionReopen  = "reopen"
	ActionCancel  = "cancel"
	ActionClaim   = "claim"
	ActionAssign  = "assign"

	ScopeAll      = "all"
	ScopeCreated  = "created"
	ScopeAssigned = "assigned"
	ScopePool     = "pool"
)

// Ticket 对应 tickets 表一行，并带上创建人 / 处理人显示名（JOIN 出来，不落库）。
type Ticket struct {
	ID           int64
	Title        string
	Description  string
	Category     string
	Status       string
	CreatorID    int64
	AssigneeID   *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     *time.Time
	CreatorName  string
	AssigneeName string
}

// PublicTicket 列表和写接口返回的工单。保留 *_id，并加上人名对象。
type PublicTicket struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Category         string     `json:"category"`
	Status           string     `json:"status"`
	CreatorID        int64      `json:"creator_id"`
	AssigneeID       *int64     `json:"assignee_id"`
	Creator          UserRef    `json:"creator"`
	Assignee         *UserRef   `json:"assignee"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ClosedAt         *time.Time `json:"closed_at"`
	AvailableActions []string   `json:"available_actions,omitempty"`
}

func (t Ticket) Public() PublicTicket {
	p := PublicTicket{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Status:      t.Status,
		CreatorID:   t.CreatorID,
		AssigneeID:  t.AssigneeID,
		Creator:     UserRef{ID: t.CreatorID, DisplayName: t.CreatorName},
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		ClosedAt:    t.ClosedAt,
	}
	if t.AssigneeID != nil {
		p.Assignee = &UserRef{ID: *t.AssigneeID, DisplayName: t.AssigneeName}
	}
	return p
}

// TicketDetail 详情：工单 + 评论 + 审计 + 当前用户能做的动作。
type TicketDetail struct {
	PublicTicket
	Comments []PublicComment `json:"comments"`
	Audits   []PublicAudit   `json:"audits"`
}

// TicketPage 列表分页信封。
type TicketPage struct {
	Items    []PublicTicket `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// TicketStats 列表页三个盒子的计数。
type TicketStats struct {
	Created  int64 `json:"created"`
	Assigned int64 `json:"assigned"`
	Pool     int64 `json:"pool"`
	All      int64 `json:"all"`
}

// TicketListQuery 列表查询。AssigneeID 仅管理员生效。
type TicketListQuery struct {
	Page       int
	PageSize   int
	Status     string
	Category   string
	Q          string
	Scope      string
	AssigneeID *int64
}

// TicketFilter 给 store 拼 WHERE。Viewer* 决定可见范围，其余是筛选。
type TicketFilter struct {
	ViewerID   int64
	ViewerRole string
	Scope      string
	Status     string
	Category   string
	Q          string
	AssigneeID *int64
	Limit      int
	Offset     int
}

// CreateTicketInput 创建工单入参。客户端就算传 status / creator_id 也会被忽略。
type CreateTicketInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// AssignTicketInput 指派入参。处理人是要派给谁，不是当前管理员。
type AssignTicketInput struct {
	AssigneeID int64 `json:"assignee_id"`
}

// TicketActionInput 改状态只传动作。reason 给重开、管理员代关、撤回用。
type TicketActionInput struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}
