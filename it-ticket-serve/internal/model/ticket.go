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
	StatusPending    = "pending"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"

	PriorityP1 = "p1"
	PriorityP2 = "p2"
	PriorityP3 = "p3"

	AuditCreate   = "create"
	AuditAssign   = "assign"
	AuditClaim    = "claim"
	AuditStart    = "start"
	AuditResolve  = "resolve"
	AuditClose    = "close"
	AuditReopen   = "reopen"
	AuditCancel   = "cancel"
	AuditWait     = "wait"
	AuditResume   = "resume"
	AuditTransfer = "transfer"
	AuditRelease  = "release"

	ActionStart    = "start"
	ActionResolve  = "resolve"
	ActionClose    = "close"
	ActionReopen   = "reopen"
	ActionCancel   = "cancel"
	ActionClaim    = "claim"
	ActionAssign   = "assign"
	ActionWait     = "wait"
	ActionResume   = "resume"
	ActionTransfer = "transfer"
	ActionEdit     = "edit"

	AuditEdit = "edit"

	ScopeAll      = "all"
	ScopeCreated  = "created"
	ScopeAssigned = "assigned"
	ScopeWaiting  = "waiting"
	ScopePool     = "pool"
)

// Ticket 对应 tickets 表一行，并带上创建人 / 处理人显示名（JOIN 出来，不落库）。
type Ticket struct {
	ID           int64
	Title        string
	Description  string
	Category     string
	Priority     string
	Status       string
	CreatorID    int64
	AssigneeID   *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     *time.Time
	CreatorName  string
	AssigneeName string

	LastCommentBody     string
	LastCommentAuthorID *int64
	LastCommentAuthor   string
	LastCommentAt       *time.Time
}

// PublicTicket 列表和写接口返回的工单。保留 *_id，并加上人名对象。
type PublicTicket struct {
	ID               int64               `json:"id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Category         string              `json:"category"`
	Priority         string              `json:"priority"`
	Status           string              `json:"status"`
	CreatorID        int64               `json:"creator_id"`
	AssigneeID       *int64              `json:"assignee_id"`
	Creator          UserRef             `json:"creator"`
	Assignee         *UserRef            `json:"assignee"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	ClosedAt         *time.Time          `json:"closed_at"`
	Stale            bool                `json:"stale"`
	LastComment      *LastCommentPreview `json:"last_comment,omitempty"`
	AvailableActions []string            `json:"available_actions,omitempty"`
}

// LastCommentPreview 列表上的最后一条评论，少进一次详情。
type LastCommentPreview struct {
	Body      string    `json:"body"`
	Author    UserRef   `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func (t Ticket) Public() PublicTicket {
	p := PublicTicket{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Priority:    t.Priority,
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
	if t.LastCommentAt != nil && t.LastCommentBody != "" {
		authorID := int64(0)
		if t.LastCommentAuthorID != nil {
			authorID = *t.LastCommentAuthorID
		}
		p.LastComment = &LastCommentPreview{
			Body:      t.LastCommentBody,
			Author:    UserRef{ID: authorID, DisplayName: t.LastCommentAuthor},
			CreatedAt: *t.LastCommentAt,
		}
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

// TicketStats 列表页盒子计数。Assigned 只含进行中的单，Waiting 是等对方的单。
type TicketStats struct {
	Created  int64 `json:"created"`
	Assigned int64 `json:"assigned"`
	Waiting  int64 `json:"waiting"`
	Pool     int64 `json:"pool"`
	All      int64 `json:"all"`
}

// TicketListQuery 列表查询。AssigneeID 仅管理员生效。
type TicketListQuery struct {
	Page       int
	PageSize   int
	Status     string
	Category   string
	Priority   string
	Q          string
	Scope      string
	AssigneeID *int64
	TimeField  string
	FromDay    string
	ToDay      string
}

// TicketFilter 给 store 拼 WHERE。Viewer* 决定可见范围，其余是筛选。
type TicketFilter struct {
	ViewerID   int64
	ViewerRole string
	Scope      string
	Status     string
	Category   string
	Priority   string
	Q          string
	AssigneeID *int64
	TimeField  string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type TicketBoard struct {
	Pool          int64       `json:"pool"`
	PoolStale     int64       `json:"pool_stale"`
	InProgress    int64       `json:"in_progress"`
	Pending       int64       `json:"pending"`
	Resolved      int64       `json:"resolved"`
	ResolvedStale int64       `json:"resolved_stale"`
	Agents        []AgentLoad `json:"agents"`
}

type AgentLoad struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Active      int64  `json:"active"`
	Waiting     int64  `json:"waiting"`
}

type CloseStaleResult struct {
	Closed int64 `json:"closed"`
}

type EditTicketInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

// CreateTicketInput 创建工单入参。客户端就算传 status / creator_id 也会被忽略。
type CreateTicketInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

// AssignTicketInput 指派入参。处理人是要派给谁，不是当前管理员。
type AssignTicketInput struct {
	AssigneeID int64 `json:"assignee_id"`
}

// TransferTicketInput 当前处理人或管理员把单交给另一位 IT。
type TransferTicketInput struct {
	AssigneeID int64  `json:"assignee_id"`
	Reason     string `json:"reason"`
}

// TicketActionInput 改状态只传动作。reason 给重开、管理员代关、撤回用。
type TicketActionInput struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}
