package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type VideoRepository struct {
	pool *pgxpool.Pool
}

func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

func (r *VideoRepository) List(ctx context.Context, search, category string, limit, offset int) ([]*domain.LongVideo, error) {
	args := []any{}
	where := "v.deleted_at IS NULL AND v.is_published = TRUE"
	if search != "" {
		args = append(args, "%"+search+"%")
		where += " AND v.title ILIKE $" + itoa(len(args))
	}
	if category != "" {
		args = append(args, category)
		where += " AND v.category = $" + itoa(len(args))
	}
	args = append(args, limit, offset)
	limOff := "LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	query := `
		SELECT v.id, v.title, v.description, v.video_url, v.thumbnail_url,
		       v.duration_sec, v.category,
		       v.views_count, v.reactions_count, v.comments_count, v.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM long_videos v
		JOIN users u ON u.id = v.author_id
		WHERE ` + where + `
		ORDER BY v.created_at DESC
		` + limOff

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.LongVideo
	for rows.Next() {
		v := &domain.LongVideo{Author: &domain.PostAuthor{}}
		var authorID uuid.UUID
		err := rows.Scan(
			&v.ID, &v.Title, &v.Description, &v.VideoURL, &v.ThumbnailURL,
			&v.DurationSec, &v.Category,
			&v.ViewsCount, &v.ReactionsCount, &v.CommentsCount, &v.CreatedAt,
			&authorID, &v.Author.Username, &v.Author.FullName, &v.Author.AvatarURL,
			&v.Author.IsVerified, &v.Author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		v.Author.ID = authorID
		list = append(list, v)
	}
	return list, rows.Err()
}

func (r *VideoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LongVideo, error) {
	const query = `
		SELECT v.id, v.title, v.description, v.video_url, v.thumbnail_url,
		       v.duration_sec, v.category,
		       v.views_count, v.reactions_count, v.comments_count, v.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM long_videos v
		JOIN users u ON u.id = v.author_id
		WHERE v.id = $1 AND v.deleted_at IS NULL
	`
	v := &domain.LongVideo{Author: &domain.PostAuthor{}}
	var authorID uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.Title, &v.Description, &v.VideoURL, &v.ThumbnailURL,
		&v.DurationSec, &v.Category,
		&v.ViewsCount, &v.ReactionsCount, &v.CommentsCount, &v.CreatedAt,
		&authorID, &v.Author.Username, &v.Author.FullName, &v.Author.AvatarURL,
		&v.Author.IsVerified, &v.Author.IsBusiness,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrVideoNotFound
	}
	if err != nil {
		return nil, err
	}
	v.Author.ID = authorID
	return v, nil
}

func (r *VideoRepository) Create(ctx context.Context, authorID uuid.UUID, in domain.CreateVideoInput) (uuid.UUID, error) {
	// Avtor faqat usta yoki sotuvchi bo'lishi kerak
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, authorID).Scan(&role)
	if err != nil {
		return uuid.Nil, err
	}
	if role != "master" && role != "seller" {
		return uuid.Nil, domain.ErrOnlyBusinessVideo
	}

	var thumb *string
	if in.ThumbnailURL != "" {
		thumb = &in.ThumbnailURL
	}
	var desc *string
	if in.Description != "" {
		desc = &in.Description
	}
	var cat *string
	if in.Category != "" {
		cat = &in.Category
	}

	var id uuid.UUID
	err = r.pool.QueryRow(ctx, `
		INSERT INTO long_videos (author_id, title, description, video_url, thumbnail_url,
		                         duration_sec, category)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, authorID, in.Title, desc, in.VideoURL, thumb, in.DurationSec, cat).Scan(&id)
	return id, err
}

func (r *VideoRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE long_videos SET views_count = views_count + 1 WHERE id = $1`, id)
	return err
}
