package model

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PublicUser 公共用户
type PublicUser struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
}

// AdminUser 管理员看用户列表用。比 PublicUser 多 created_at，仍不含密码。
type AdminUser struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u User) AdminPublic() AdminUser {
	return AdminUser{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
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

// RegisterInput 注册输入
type RegisterInput struct {
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
