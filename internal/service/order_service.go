package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type OrderService struct {
	repo *postgres.OrderRepository
}

func NewOrderService(repo *postgres.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Create(ctx context.Context, userID uuid.UUID, in domain.CreateOrderInput) (*domain.Order, error) {
	return s.repo.CreateFromCart(ctx, userID, in)
}

func (s *OrderService) List(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	list, err := s.repo.List(ctx, userID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Order{}
	}
	return list, nil
}

func (s *OrderService) Cancel(ctx context.Context, orderID, userID uuid.UUID) error {
	return s.repo.Cancel(ctx, orderID, userID)
}
