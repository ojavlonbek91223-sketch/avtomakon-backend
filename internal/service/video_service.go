package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type VideoService struct {
	repo *postgres.VideoRepository
}

func NewVideoService(repo *postgres.VideoRepository) *VideoService {
	return &VideoService{repo: repo}
}

func (s *VideoService) List(ctx context.Context, search, category string, limit, offset int) ([]*domain.LongVideo, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	list, err := s.repo.List(ctx, search, category, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.LongVideo{}
	}
	return list, nil
}

func (s *VideoService) Get(ctx context.Context, id uuid.UUID) (*domain.LongVideo, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err == nil {
		go s.repo.IncrementViews(context.Background(), id)
	}
	return v, err
}

func (s *VideoService) Create(ctx context.Context, authorID uuid.UUID, in domain.CreateVideoInput) (uuid.UUID, error) {
	return s.repo.Create(ctx, authorID, in)
}
