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
