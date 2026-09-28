package model

import "time"

// UserRef 工单 / 评论 / 审计里带出的人名。只放 id 和显示名，不含邮箱和角色。
type UserRef struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
}

const (
	UserActive   = "active"
	UserDisabled = "disabled"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u User) Disabled() bool {
	return u.Status == UserDisabled
}

// PublicUser 公共用户
type PublicUser struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	UnreadCount int64  `json:"unread_count"`
}

func (u User) Public() PublicUser {
	status := u.Status
	if status == "" {
		status = UserActive
	}
	return PublicUser{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Status:      status,
	}
}

// AdminUser 管理员看用户列表用。比 PublicUser 多 created_at，仍不含密码。
type AdminUser struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	DisplayName     string    `json:"display_name"`
	Role            string    `json:"role"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	OpenTicketCount int64     `json:"open_ticket_count"`
	ReleasedCount   int64     `json:"released_count,omitempty"`
}

func (u User) AdminPublic() AdminUser {
	status := u.Status
	if status == "" {
		status = UserActive
	}
	return AdminUser{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Status:      status,
		CreatedAt:   u.CreatedAt,
	}
}

// UserPage 用户列表的分页信封，放在统一响应的 data 里。
type UserPage struct {
	Items    []AdminUser `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// UpdateRoleInput 改用户角色。只允许 user / agent / admin。
type UpdateRoleInput struct {
	Role string `json:"role"`
}

// UpdateUserStatusInput 停用 / 启用账号。
type UpdateUserStatusInput struct {
	Status string `json:"status"`
}

// RegisterInput 注册输入
type RegisterInput struct {
	InviteCode  string `json:"invite_code"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginInput 登录输入
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResult 登录结果
type LoginResult struct {
	Token string     `json:"token"`
	User  PublicUser `json:"user"`
}

// UpdateProfileInput 自己改显示名。
type UpdateProfileInput struct {
	DisplayName string `json:"display_name"`
}

// UpdatePasswordInput 自己改密码，必须带上当前密码。
type UpdatePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
