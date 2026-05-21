package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type BusinessApplicationService struct {
	repo *postgres.BusinessApplicationRepository
}

func NewBusinessApplicationService(repo *postgres.BusinessApplicationRepository) *BusinessApplicationService {
	return &BusinessApplicationService{repo: repo}
}

func (s *BusinessApplicationService) Apply(ctx context.Context, userID uuid.UUID, in domain.ApplyBusinessInput) (*domain.BusinessApplication, error) {
	// Bir vaqtning o'zida bittadan ortiq ariza qabul qilmaymiz
	pending, err := s.repo.HasPending(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, domain.ErrApplicationAlreadyPending
	}

	var description *string
	if d := strings.TrimSpace(in.Description); d != "" {
		description = &d
	}

	photos := in.WorkplacePhotos
	if photos == nil {
		photos = []string{}
	}

	return s.repo.Create(ctx, postgres.CreateApplicationInput{
		UserID:          userID,
		Type:            in.Type,
		BusinessName:    in.BusinessName,
		ContactPhone:    in.ContactPhone,
		Address:         in.Address,
		Lat:             in.Location.Lat,
		Lng:             in.Location.Lng,
		ExperienceYears: in.ExperienceYears,
		Description:     description,
		WorkplacePhotos: photos,
		DocumentURL:     in.DocumentURL,
	})
}

func (s *BusinessApplicationService) Mine(ctx context.Context, userID uuid.UUID) ([]*domain.BusinessApplication, error) {
	apps, err := s.repo.FindMine(ctx, userID)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		apps = []*domain.BusinessApplication{}
	}
	return apps, nil
}
