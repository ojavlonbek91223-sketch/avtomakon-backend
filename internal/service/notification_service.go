package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
	ws "github.com/avtomakon/backend/internal/websocket"
)

type NotificationService struct {
	repo *postgres.NotificationRepository
	hub  *ws.Hub
}

func NewNotificationService(repo *postgres.NotificationRepository, hub *ws.Hub) *NotificationService {
	return &NotificationService{repo: repo, hub: hub}
}

func (s *NotificationService) Create(ctx context.Context, in postgres.CreateNotificationInput) (*domain.Notification, error) {
	// O'ziga bildirishnoma yubormaymiz
	if in.ActorID != nil && *in.ActorID == in.UserID {
		return nil, nil
	}

	n, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	// WS bilan real-time
	s.hub.SendToUser(in.UserID, ws.Event{
		Event: "notification.new",
		Data:  map[string]any{"notification": n},
	})

	return n, nil
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit int) ([]*domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	list, err := s.repo.List(ctx, userID, unreadOnly, limit)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Notification{}
	}
	return list, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) SavePushToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	return s.repo.SavePushToken(ctx, userID, token, platform)
}
