package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type UserService struct {
	repo  *postgres.UserRepository
	notif *NotificationService
}

func NewUserService(repo *postgres.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SetNotificationService(n *NotificationService) {
	s.notif = n
}

func (s *UserService) Me(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, in postgres.UpdateProfileInput) (*domain.User, error) {
	return s.repo.UpdateProfile(ctx, id, in)
}

func (s *UserService) GetPublic(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.User, bool, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, false, err
	}
	following := false
	if viewerID != nil && *viewerID != id {
		following, _ = s.repo.IsFollowing(ctx, *viewerID, id)
	}
	return user, following, nil
}

func (s *UserService) Follow(ctx context.Context, follower, target uuid.UUID) (int, error) {
	count, err := s.repo.Follow(ctx, follower, target)
	if err != nil {
		return 0, err
	}
	if s.notif != nil {
		entityType := "user"
		_, _ = s.notif.Create(ctx, postgres.CreateNotificationInput{
			UserID:     target,
			Type:       domain.NotificationFollow,
			ActorID:    &follower,
			EntityType: &entityType,
			EntityID:   &follower,
			Title:      "Sizni kuzata boshladi",
		})
	}
	return count, nil
}

func (s *UserService) Unfollow(ctx context.Context, follower, target uuid.UUID) (int, error) {
	return s.repo.Unfollow(ctx, follower, target)
}

func (s *UserService) Followers(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, limit, offset int) ([]*domain.UserBrief, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	list, err := s.repo.Followers(ctx, userID, viewerID, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.UserBrief{}
	}
	return list, nil
}

func (s *UserService) Following(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, limit, offset int) ([]*domain.UserBrief, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	list, err := s.repo.Following(ctx, userID, viewerID, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.UserBrief{}
	}
	return list, nil
}
