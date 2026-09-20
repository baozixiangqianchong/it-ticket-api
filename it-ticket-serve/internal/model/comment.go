package model

import "time"

// Comment 对应 ticket_comments 表一行。
type Comment struct {
	ID        int64
	TicketID  int64
	AuthorID  int64
	Body      string
	CreatedAt time.Time
}

// PublicComment 接口返回的评论。
type PublicComment struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	AuthorID  int64     `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Comment) Public() PublicComment {
	return PublicComment{
		ID:        c.ID,
		TicketID:  c.TicketID,
		AuthorID:  c.AuthorID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
}

// CreateCommentInput 追加评论。作者由当前登录用户决定，不能客户端指定。
type CreateCommentInput struct {
	Body string `json:"body"`
}
