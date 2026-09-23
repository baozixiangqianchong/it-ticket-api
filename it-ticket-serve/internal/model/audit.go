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
