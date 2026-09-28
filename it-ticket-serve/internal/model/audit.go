package model

import "time"

// Audit 对应 audit_logs 一行。ActorName 是 JOIN 出来的显示名。
type Audit struct {
	ID         int64
	TicketID   int64
	ActorID    int64
	Action     string
	FromStatus *string
	ToStatus   string
	Reason     *string
	CreatedAt  time.Time
	ActorName  string
}

// PublicAudit 详情时间线。
type PublicAudit struct {
	ID         int64     `json:"id"`
	Action     string    `json:"action"`
	FromStatus *string   `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Reason     *string   `json:"reason"`
	Actor      UserRef   `json:"actor"`
	CreatedAt  time.Time `json:"created_at"`
}

func (a Audit) Public() PublicAudit {
	return PublicAudit{
		ID:         a.ID,
		Action:     a.Action,
		FromStatus: a.FromStatus,
		ToStatus:   a.ToStatus,
		Reason:     a.Reason,
		Actor:      UserRef{ID: a.ActorID, DisplayName: a.ActorName},
		CreatedAt:  a.CreatedAt,
	}
}

const (
	ActivityKindTicket  = "ticket"
	ActivityKindAccount = "account"

	AccountActionRole   = "role"
	AccountActionStatus = "status"
)

// Activity 管理员全局审计一行：工单时间线或账号变更。
type Activity struct {
	ID          int64
	Kind        string
	Action      string
	ActorID     int64
	ActorName   string
	TicketID    *int64
	TicketTitle *string
	UserID      *int64
	UserName    *string
	FromValue   *string
	ToValue     string
	Reason      *string
	CreatedAt   time.Time
}

type PublicActivity struct {
	ID          int64     `json:"id"`
	Kind        string    `json:"kind"`
	Action      string    `json:"action"`
	Actor       UserRef   `json:"actor"`
	TicketID    *int64    `json:"ticket_id"`
	TicketTitle *string   `json:"ticket_title"`
	UserID      *int64    `json:"user_id"`
	UserName    *string   `json:"user_name"`
	FromValue   *string   `json:"from_value"`
	ToValue     string    `json:"to_value"`
	Reason      *string   `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

func (a Activity) Public() PublicActivity {
	return PublicActivity{
		ID:          a.ID,
		Kind:        a.Kind,
		Action:      a.Action,
		Actor:       UserRef{ID: a.ActorID, DisplayName: a.ActorName},
		TicketID:    a.TicketID,
		TicketTitle: a.TicketTitle,
		UserID:      a.UserID,
		UserName:    a.UserName,
		FromValue:   a.FromValue,
		ToValue:     a.ToValue,
		Reason:      a.Reason,
		CreatedAt:   a.CreatedAt,
	}
}

type ActivityPage struct {
	Items    []PublicActivity `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type ActivityFilter struct {
	Kind   string
	Action string
	Q      string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}
