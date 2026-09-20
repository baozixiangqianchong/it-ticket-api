package service

import (
	"database/sql"
	"errors"
	"strings"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

// AdminService 仅管理员使用的账号管理。路由上还会再挡一层 RequireAdmin。
type AdminService struct {
	users *store.UserStore
}

func NewAdminService(users *store.UserStore) *AdminService {
	return &AdminService{users: users}
}

// ListUsers 分页用户列表。不含密码哈希。
func (s *AdminService) ListUsers(page, pageSize int) (*model.UserPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	total, err := s.users.Count()
	if err != nil {
		return nil, err
	}
	list, err := s.users.List((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.AdminUser, 0, len(list))
	for _, u := range list {
		items = append(items, u.AdminPublic())
	}
	return &model.UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// UpdateRole 把用户改成 user / agent / admin。不能拿掉系统里最后一个 admin。
func (s *AdminService) UpdateRole(userID int64, role string) (*model.AdminUser, error) {
	role = strings.TrimSpace(role)
	if role != "user" && role != "agent" && role != "admin" {
		return nil, invalidArg("角色必须是 user / agent / admin")
	}

	u, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("用户不存在")
		}
		return nil, err
	}

	// 最后一个管理员降级后，没人能派单、也没人能再改角色。
	if u.Role == "admin" && role != "admin" {
		n, err := s.users.CountByRole("admin")
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			return nil, conflict("不能取消最后一个管理员")
		}
	}

	if u.Role != role {
		if err := s.users.UpdateRole(userID, role); err != nil {
			return nil, err
		}
		u.Role = role
	}
	pub := u.AdminPublic()
	return &pub, nil
}
