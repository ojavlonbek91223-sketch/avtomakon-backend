package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type CartService struct {
	repo  *postgres.CartRepository
	order *postgres.OrderRepository
}

func NewCartService(repo *postgres.CartRepository, order *postgres.OrderRepository) *CartService {
	return &CartService{repo: repo, order: order}
}

func (s *CartService) Get(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	return s.repo.Get(ctx, userID)
}

func (s *CartService) Add(ctx context.Context, userID uuid.UUID, in domain.AddToCartInput) error {
	productID, err := uuid.Parse(in.ProductID)
	if err != nil {
		return errors.New("noto'g'ri product_id")
	}
	return s.repo.AddItem(ctx, userID, productID, in.Quantity)
}

func (s *CartService) Update(ctx context.Context, userID, itemID uuid.UUID, quantity int) error {
	return s.repo.UpdateItem(ctx, userID, itemID, quantity)
}

func (s *CartService) Remove(ctx context.Context, userID, itemID uuid.UUID) error {
	return s.repo.RemoveItem(ctx, userID, itemID)
}

func (s *CartService) Clear(ctx context.Context, userID uuid.UUID) error {
	return s.repo.Clear(ctx, userID)
}
