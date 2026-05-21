package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type ReviewService struct {
	repo *postgres.ReviewRepository
}

func NewReviewService(repo *postgres.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) Create(ctx context.Context, authorID, targetID uuid.UUID, targetType domain.ReviewTargetType, in domain.CreateReviewInput) (*domain.Review, error) {
	var text *string
	if in.Text != "" {
		text = &in.Text
	}

	var orderID *uuid.UUID
	if in.OrderID != nil && *in.OrderID != "" {
		oid, err := uuid.Parse(*in.OrderID)
		if err == nil {
			orderID = &oid
		}
	}

	images := in.Images
	if images == nil {
		images = []string{}
	}

	return s.repo.Create(ctx, postgres.CreateReviewDBInput{
		AuthorID:   authorID,
		TargetType: targetType,
		TargetID:   targetID,
		OrderID:    orderID,
		Rating:     in.Rating,
		Text:       text,
		Images:     images,
	})
}

func (s *ReviewService) ListByTarget(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID, limit, offset int) ([]*domain.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	list, err := s.repo.ListByTarget(ctx, targetType, targetID, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Review{}
	}
	return list, nil
}

func (s *ReviewService) MyReviews(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	list, err := s.repo.ListByAuthor(ctx, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Review{}
	}
	return list, nil
}

func (s *ReviewService) Reply(ctx context.Context, reviewID, ownerID uuid.UUID, text string) error {
	return s.repo.Reply(ctx, reviewID, ownerID, text)
}

func (s *ReviewService) Delete(ctx context.Context, reviewID, authorID uuid.UUID) error {
	return s.repo.Delete(ctx, reviewID, authorID)
}
