package service

import (
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

type NotificationService struct {
	notifs *store.NotificationStore
}

func NewNotificationService(notifs *store.NotificationStore) *NotificationService {
	return &NotificationService{notifs: notifs}
}

func (s *NotificationService) List(userID int64, unreadOnly bool, page, pageSize int) (*model.NotificationPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	total, err := s.notifs.Count(userID, unreadOnly)
	if err != nil {
		return nil, err
	}
	rows, err := s.notifs.List(userID, unreadOnly, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]model.PublicNotification, 0, len(rows))
	for _, n := range rows {
		items = append(items, n.Public())
	}
	return &model.NotificationPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *NotificationService) MarkRead(userID, id int64) error {
	ok, err := s.notifs.Owned(userID, id)
	if err != nil {
		return err
	}
	if !ok {
		return notFound("通知不存在")
	}
	_, err = s.notifs.MarkRead(userID, id)
	return err
}

func (s *NotificationService) MarkAllRead(userID int64) error {
	return s.notifs.MarkAllRead(userID)
}

func (s *NotificationService) UnreadCount(userID int64) (int64, error) {
	return s.notifs.UnreadCount(userID)
}
