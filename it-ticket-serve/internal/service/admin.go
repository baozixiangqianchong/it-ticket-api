package service

import (
	"database/sql"
	"errors"
	"strings"

	"it-ticket-api/internal/logger"
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

func (s *AdminService) ListUsers(page, pageSize int, role, q string) (*model.UserPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	role = strings.TrimSpace(role)
	if role != "" && role != "user" && role != "agent" && role != "admin" {
		return nil, invalidArg("角色必须是 user / agent / admin")
	}

	total, err := s.users.Count(role, q)
	if err != nil {
		return nil, err
	}
	list, err := s.users.List((page-1)*pageSize, pageSize, role, q)
	if err != nil {
		return nil, err
	}

	items := make([]model.AdminUser, 0, len(list))
	for _, u := range list {
		items = append(items, u.AdminPublic())
	}
	return &model.UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) UpdateRole(actorID, userID int64, role string) (*model.AdminUser, error) {
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

	// 不能拿掉自己的管理员：改成功后 JWT 下一请求就进不了 /admin，页面会误报「没有权限」。
	if actorID == userID && u.Role == "admin" && role != "admin" {
		return nil, permissionDenied("不能取消自己的管理员身份")
	}

	if u.Role == "admin" && role != "admin" {
		n, err := s.users.CountByRole("admin")
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			return nil, lastAdmin()
		}
	}

	if u.Role != role {
		if err := s.users.UpdateRole(userID, role); err != nil {
			return nil, err
		}
		logger.Info("用户角色已变更", "user_id", userID, "from", u.Role, "to", role)
		u.Role = role
	}
	pub := u.AdminPublic()
	return &pub, nil
}
