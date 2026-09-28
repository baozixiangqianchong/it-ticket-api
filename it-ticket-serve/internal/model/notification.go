package model

import "time"

const (
	NotifyAssign   = "assign"
	NotifyClaim    = "claim"
	NotifyStart    = "start"
	NotifyWait     = "wait"
	NotifyResume   = "resume"
	NotifyResolve  = "resolve"
	NotifyClose    = "close"
	NotifyReopen   = "reopen"
	NotifyComment  = "comment"
	NotifyTransfer = "transfer"
	NotifyCancel   = "cancel"
	NotifyRole     = "role"
	NotifyAccount  = "account"
	NotifyRelease  = "release"
)

type Notification struct {
	ID        int64
	UserID    int64
	TicketID  *int64
	Type      string
	Title     string
	Body      string
	ReadAt    *time.Time
	CreatedAt time.Time
}

type PublicNotification struct {
	ID        int64     `json:"id"`
	TicketID  *int64    `json:"ticket_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

func (n Notification) Public() PublicNotification {
	return PublicNotification{
		ID:        n.ID,
		TicketID:  n.TicketID,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.ReadAt != nil,
		CreatedAt: n.CreatedAt,
	}
}

type NotificationPage struct {
	Items    []PublicNotification `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type AgentRef struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type AgentList struct {
	Items []AgentRef `json:"items"`
}
