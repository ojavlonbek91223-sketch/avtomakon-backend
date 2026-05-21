package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type ReviewRepository struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

type CreateReviewDBInput struct {
	AuthorID   uuid.UUID
	TargetType domain.ReviewTargetType
	TargetID   uuid.UUID
	OrderID    *uuid.UUID
	Rating     int
	Text       *string
	Images     []string
}

func (r *ReviewRepository) Create(ctx context.Context, in CreateReviewDBInput) (*domain.Review, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rev := &domain.Review{
		Author:     &domain.PostAuthor{ID: in.AuthorID},
		TargetType: in.TargetType,
		TargetID:   in.TargetID,
		OrderID:    in.OrderID,
		Rating:     in.Rating,
		Text:       in.Text,
		Images:     in.Images,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO reviews (author_id, target_type, target_id, order_id, rating, text, images)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, in.AuthorID, in.TargetType, in.TargetID, in.OrderID, in.Rating, in.Text, in.Images).
		Scan(&rev.ID, &rev.CreatedAt)
	if err != nil {
		if isUniqueViolation(err, "") {
			return nil, domain.ErrReviewExists
		}
		return nil, err
	}

	// Reyting yangilash (faqat business va product uchun)
	if in.TargetType == domain.ReviewTargetBusiness {
		_, err = tx.Exec(ctx, `
			UPDATE businesses SET
				rating_count = rating_count + 1,
				rating_avg = ROUND((rating_avg * rating_count + $1) / (rating_count + 1)::numeric, 1)
			WHERE id = $2
		`, in.Rating, in.TargetID)
		if err != nil {
			return nil, err
		}
	} else if in.TargetType == domain.ReviewTargetProduct {
		_, err = tx.Exec(ctx, `
			UPDATE products SET
				rating_count = rating_count + 1,
				rating_avg = ROUND((rating_avg * rating_count + $1) / (rating_count + 1)::numeric, 1)
			WHERE id = $2
		`, in.Rating, in.TargetID)
		if err != nil {
			return nil, err
		}
	}

	// Author ma'lumotlari
	err = tx.QueryRow(ctx, `
		SELECT username, full_name, avatar_url, is_verified, is_business
		FROM users WHERE id = $1
	`, in.AuthorID).Scan(
		&rev.Author.Username, &rev.Author.FullName, &rev.Author.AvatarURL,
		&rev.Author.IsVerified, &rev.Author.IsBusiness,
	)
	if err != nil {
		return nil, err
	}

	return rev, tx.Commit(ctx)
}

func (r *ReviewRepository) ListByTarget(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID, limit, offset int) ([]*domain.Review, error) {
	const query = `
		SELECT r.id, r.target_type, r.target_id, r.order_id, r.rating, r.text, r.images,
		       r.seller_reply, r.seller_replied_at, r.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM reviews r
		JOIN users u ON u.id = r.author_id AND u.deleted_at IS NULL
		WHERE r.target_type = $1 AND r.target_id = $2
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4
	`
	return r.scanList(ctx, query, targetType, targetID, limit, offset)
}

func (r *ReviewRepository) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Review, error) {
	const query = `
		SELECT r.id, r.target_type, r.target_id, r.order_id, r.rating, r.text, r.images,
		       r.seller_reply, r.seller_replied_at, r.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business,
		       CASE
		         WHEN r.target_type = 'business' THEN (SELECT name FROM businesses WHERE id = r.target_id)
		         WHEN r.target_type = 'product' THEN (SELECT name FROM products WHERE id = r.target_id)
		       END AS target_name
		FROM reviews r
		JOIN users u ON u.id = r.author_id
		WHERE r.author_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Review
	for rows.Next() {
		rev := &domain.Review{Author: &domain.PostAuthor{}}
		err := rows.Scan(
			&rev.ID, &rev.TargetType, &rev.TargetID, &rev.OrderID,
			&rev.Rating, &rev.Text, &rev.Images,
			&rev.SellerReply, &rev.SellerRepliedAt, &rev.CreatedAt,
			&rev.Author.ID, &rev.Author.Username, &rev.Author.FullName,
			&rev.Author.AvatarURL, &rev.Author.IsVerified, &rev.Author.IsBusiness,
			&rev.TargetName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, rev)
	}
	return list, rows.Err()
}

func (r *ReviewRepository) scanList(ctx context.Context, query string, args ...any) ([]*domain.Review, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Review
	for rows.Next() {
		rev := &domain.Review{Author: &domain.PostAuthor{}}
		err := rows.Scan(
			&rev.ID, &rev.TargetType, &rev.TargetID, &rev.OrderID,
			&rev.Rating, &rev.Text, &rev.Images,
			&rev.SellerReply, &rev.SellerRepliedAt, &rev.CreatedAt,
			&rev.Author.ID, &rev.Author.Username, &rev.Author.FullName,
			&rev.Author.AvatarURL, &rev.Author.IsVerified, &rev.Author.IsBusiness,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, rev)
	}
	return list, rows.Err()
}

func (r *ReviewRepository) Reply(ctx context.Context, reviewID, businessOwnerID uuid.UUID, text string) error {
	const query = `
		UPDATE reviews SET seller_reply = $3, seller_replied_at = NOW()
		WHERE id = $1
		  AND target_type = 'business'
		  AND target_id IN (SELECT id FROM businesses WHERE owner_id = $2)
	`
	cmd, err := r.pool.Exec(ctx, query, reviewID, businessOwnerID, text)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("sharh topilmadi yoki siz biznes egasi emassiz")
	}
	return nil
}

func (r *ReviewRepository) Delete(ctx context.Context, reviewID, authorID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var targetType domain.ReviewTargetType
	var targetID uuid.UUID
	var rating int
	err = tx.QueryRow(ctx, `
		DELETE FROM reviews WHERE id = $1 AND author_id = $2
		RETURNING target_type, target_id, rating
	`, reviewID, authorID).Scan(&targetType, &targetID, &rating)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("sharh topilmadi")
	}
	if err != nil {
		return err
	}

	// Reytingni qaytarib hisoblash
	switch targetType {
	case domain.ReviewTargetBusiness:
		_, err = tx.Exec(ctx, `
			UPDATE businesses SET
				rating_count = GREATEST(rating_count - 1, 0),
				rating_avg = CASE
					WHEN rating_count - 1 <= 0 THEN 0
					ELSE ROUND((rating_avg * rating_count - $1) / (rating_count - 1)::numeric, 1)
				END
			WHERE id = $2
		`, rating, targetID)
	case domain.ReviewTargetProduct:
		_, err = tx.Exec(ctx, `
			UPDATE products SET
				rating_count = GREATEST(rating_count - 1, 0),
				rating_avg = CASE
					WHEN rating_count - 1 <= 0 THEN 0
					ELSE ROUND((rating_avg * rating_count - $1) / (rating_count - 1)::numeric, 1)
				END
			WHERE id = $2
		`, rating, targetID)
	}
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
