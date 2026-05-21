package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type CommentService struct {
	repo *postgres.CommentRepository
}

func NewCommentService(repo *postgres.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) List(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, limit int) ([]*domain.Comment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	list, err := s.repo.List(ctx, postID, viewerID, limit)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*domain.Comment{}
	}
	return list, nil
}

func (s *CommentService) Create(ctx context.Context, postID, userID uuid.UUID, in domain.CreateCommentInput) (*domain.Comment, error) {
	return s.repo.Create(ctx, postID, userID, in)
}

func (s *CommentService) Delete(ctx context.Context, commentID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, commentID, userID)
}

func (s *CommentService) ToggleLike(ctx context.Context, commentID, userID uuid.UUID) (bool, int, error) {
	return s.repo.ToggleLike(ctx, commentID, userID)
}
