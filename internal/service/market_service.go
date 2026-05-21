package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type MarketService struct {
	repo *postgres.MarketRepository
}

func NewMarketService(repo *postgres.MarketRepository) *MarketService {
	return &MarketService{repo: repo}
}

func (s *MarketService) Categories(ctx context.Context) ([]*domain.Category, error) {
	list, err := s.repo.Categories(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Category{}
	}
	return list, nil
}

func (s *MarketService) ListProducts(ctx context.Context, p domain.ProductsParams) ([]*domain.Product, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	list, err := s.repo.ListProducts(ctx, p)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Product{}
	}
	return list, nil
}

func (s *MarketService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.FindProduct(ctx, id)
}

func (s *MarketService) Featured(ctx context.Context, limit int) ([]*domain.Product, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	list, err := s.repo.Featured(ctx, limit)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Product{}
	}
	return list, nil
}

func (s *MarketService) Promotions(ctx context.Context) ([]*domain.Promotion, error) {
	list, err := s.repo.ActivePromotions(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Promotion{}
	}
	return list, nil
}
