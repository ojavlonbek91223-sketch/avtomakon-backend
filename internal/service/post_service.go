package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/domain"
	"github.com/avtomakon/backend/internal/repository/postgres"
)

type PostService struct {
	repo  *postgres.PostRepository
	notif *NotificationService
}

func NewPostService(repo *postgres.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) SetNotificationService(n *NotificationService) {
	s.notif = n
}

func (s *PostService) Create(ctx context.Context, authorID uuid.UUID, in domain.CreatePostInput) (uuid.UUID, error) {
	visibility := in.Visibility
	if visibility == "" {
		visibility = domain.VisibilityPublic
	}

	// Caption (NULL bo'lishi mumkin)
	var caption *string
	if c := strings.TrimSpace(in.Caption); c != "" {
		caption = &c
	}

	// Cover — birinchi medianing thumb yoki URL
	var cover *string
	if len(in.Media) > 0 {
		if t := in.Media[0].ThumbnailURL; t != nil && *t != "" {
			cover = t
		} else {
			cover = &in.Media[0].URL
		}
	}

	// Hashtags — pastki harf va # belgisi olib tashlanadi
	tags := make([]string, 0, len(in.Hashtags))
	for _, t := range in.Hashtags {
		t = strings.TrimSpace(t)
		t = strings.TrimPrefix(t, "#")
		t = strings.ToLower(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	dbIn := postgres.CreatePostDBInput{
		AuthorID:   authorID,
		Caption:    caption,
		MediaType:  in.MediaType,
		CoverURL:   cover,
		Visibility: visibility,
		Media:      in.Media,
		Hashtags:   tags,
	}

	if in.Location != nil {
		dbIn.LocationName = &in.Location.Name
		dbIn.LocationLat = &in.Location.Lat
		dbIn.LocationLng = &in.Location.Lng
	}

	return s.repo.Create(ctx, dbIn)
}

func (s *PostService) ListFeed(ctx context.Context, kind domain.FeedKind, viewerID *uuid.UUID, cursor *time.Time, limit int) (*domain.FeedResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	posts, err := s.repo.ListFeed(ctx, kind, viewerID, cursor, limit)
	if err != nil {
		return nil, err
	}

	result := &domain.FeedResult{Posts: posts}
	if len(posts) > limit {
		result.HasMore = true
		result.Posts = posts[:limit]
		next := posts[limit-1].CreatedAt
		result.NextCursor = &next
	}

	return result, nil
}

func (s *PostService) ListSaved(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int) (*domain.FeedResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 21
	}

	posts, err := s.repo.ListSaved(ctx, userID, cursor, limit)
	if err != nil {
		return nil, err
	}

	result := &domain.FeedResult{Posts: posts}
	if len(posts) > limit {
		result.HasMore = true
		result.Posts = posts[:limit]
		next := posts[limit-1].CreatedAt
		result.NextCursor = &next
	}

	return result, nil
}

func (s *PostService) ListByUser(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, cursor *time.Time, limit int) (*domain.FeedResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 21
	}

	posts, err := s.repo.ListByUser(ctx, userID, viewerID, cursor, limit)
	if err != nil {
		return nil, err
	}

	result := &domain.FeedResult{Posts: posts}
	if len(posts) > limit {
		result.HasMore = true
		result.Posts = posts[:limit]
		next := posts[limit-1].CreatedAt
		result.NextCursor = &next
	}

	return result, nil
}

func (s *PostService) Delete(ctx context.Context, postID, authorID uuid.UUID) error {
	return s.repo.Delete(ctx, postID, authorID)
}

func (s *PostService) SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction domain.ReactionType) (*domain.ReactionResult, error) {
	res, err := s.repo.SetReaction(ctx, postID, userID, reaction)
	if err != nil {
		return nil, err
	}
	// Bildirishnoma (post egasiga)
	if s.notif != nil {
		go s.notifyOnReaction(postID, userID)
	}
	return res, nil
}

func (s *PostService) RemoveReaction(ctx context.Context, postID, userID uuid.UUID) (*domain.ReactionResult, error) {
	return s.repo.RemoveReaction(ctx, postID, userID)
}

func (s *PostService) notifyOnReaction(postID, actorID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	owner, err := s.repo.PostAuthor(ctx, postID)
	if err != nil || owner == uuid.Nil {
		return
	}
	entityType := "post"
	_, _ = s.notif.Create(ctx, postgres.CreateNotificationInput{
		UserID:     owner,
		Type:       domain.NotificationLike,
		ActorID:    &actorID,
		EntityType: &entityType,
		EntityID:   &postID,
		Title:      "Sizning postingizni yoqtirdi",
	})
}

func (s *PostService) Save(ctx context.Context, postID, userID uuid.UUID) error {
	return s.repo.Save(ctx, postID, userID)
}

func (s *PostService) Unsave(ctx context.Context, postID, userID uuid.UUID) error {
	return s.repo.Unsave(ctx, postID, userID)
}
