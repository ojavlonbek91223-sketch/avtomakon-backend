package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type BusinessService struct {
	repo *postgres.BusinessRepository
}

func NewBusinessService(repo *postgres.BusinessRepository) *BusinessService {
	return &BusinessService{repo: repo}
}

func (s *BusinessService) Nearby(ctx context.Context, p domain.NearbyParams) ([]*domain.Business, error) {
	if p.RadiusM <= 0 || p.RadiusM > 100_000 {
		p.RadiusM = 10_000 // 10 km
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 50
	}
	return s.repo.Nearby(ctx, p)
}

func (s *BusinessService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Business, error) {
	return s.repo.FindByID(ctx, id)
}
