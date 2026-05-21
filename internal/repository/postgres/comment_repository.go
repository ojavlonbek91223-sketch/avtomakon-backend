package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// List — postning izohlarini olish. viewerID berilsa, viewer_liked ham yuklanadi.
func (r *CommentRepository) List(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, limit int) ([]*domain.Comment, error) {
	args := []any{postID, limit}
	viewerSql := "FALSE"
	if viewerID != nil {
		args = append(args, *viewerID)
		viewerSql = "EXISTS(SELECT 1 FROM comment_likes WHERE user_id = $3 AND comment_id = c.id)"
	}

	query := `
		SELECT c.id, c.post_id, c.parent_id, c.text, c.likes_count, c.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business,
		       (SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL),
		       ` + viewerSql + `
		FROM comments c
		JOIN users u ON u.id = c.user_id AND u.deleted_at IS NULL
		WHERE c.post_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		c := &domain.Comment{Author: &domain.PostAuthor{}}
		var authorID uuid.UUID
		err := rows.Scan(
			&c.ID, &c.PostID, &c.ParentID, &c.Text, &c.LikesCount, &c.CreatedAt,
			&authorID, &c.Author.Username, &c.Author.FullName,
			&c.Author.AvatarURL, &c.Author.IsVerified, &c.Author.IsBusiness,
			&c.RepliesCount, &c.ViewerLiked,
		)
		if err != nil {
			return nil, err
		}
		c.Author.ID = authorID
		comments = append(comments, c)
	}

	return comments, rows.Err()
}

func (r *CommentRepository) Create(ctx context.Context, postID, userID uuid.UUID, in domain.CreateCommentInput) (*domain.Comment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var parentID *uuid.UUID
	if in.ParentID != nil && *in.ParentID != "" {
		pid, err := uuid.Parse(*in.ParentID)
		if err != nil {
			return nil, errors.New("noto'g'ri parent_id")
		}
		parentID = &pid
	}

	// Post mavjudligini tekshirish
	var postExists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND deleted_at IS NULL)`,
		postID).Scan(&postExists)
	if err != nil {
		return nil, err
	}
	if !postExists {
		return nil, domain.ErrPostNotFound
	}

	c := &domain.Comment{
		PostID: postID,
		ParentID: parentID,
		Text: in.Text,
		Author: &domain.PostAuthor{ID: userID},
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO comments (post_id, user_id, parent_id, text)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, postID, userID, parentID, in.Text).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Post counter
	_, err = tx.Exec(ctx,
		`UPDATE posts SET comments_count = comments_count + 1 WHERE id = $1`, postID)
	if err != nil {
		return nil, err
	}

	// Author ma'lumotlarini yuklash
	err = tx.QueryRow(ctx, `
		SELECT username, full_name, avatar_url, is_verified, is_business
		FROM users WHERE id = $1
	`, userID).Scan(
		&c.Author.Username, &c.Author.FullName, &c.Author.AvatarURL,
		&c.Author.IsVerified, &c.Author.IsBusiness,
	)
	if err != nil {
		return nil, err
	}

	return c, tx.Commit(ctx)
}

func (r *CommentRepository) Delete(ctx context.Context, commentID, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var postID uuid.UUID
	err = tx.QueryRow(ctx, `
		UPDATE comments SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING post_id
	`, commentID, userID).Scan(&postID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("izoh topilmadi yoki sizniki emas")
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE posts SET comments_count = GREATEST(comments_count - 1, 0) WHERE id = $1`,
		postID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *CommentRepository) ToggleLike(ctx context.Context, commentID, userID uuid.UUID) (liked bool, count int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback(ctx)

	// Mavjudligini tekshirish
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = $1 AND comment_id = $2)`,
		userID, commentID).Scan(&exists)
	if err != nil {
		return false, 0, err
	}

	if exists {
		_, err = tx.Exec(ctx,
			`DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2`,
			userID, commentID)
		if err != nil {
			return false, 0, err
		}
		_, err = tx.Exec(ctx,
			`UPDATE comments SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = $1`,
			commentID)
		liked = false
	} else {
		_, err = tx.Exec(ctx,
			`INSERT INTO comment_likes (user_id, comment_id) VALUES ($1, $2)`,
			userID, commentID)
		if err != nil {
			return false, 0, err
		}
		_, err = tx.Exec(ctx,
			`UPDATE comments SET likes_count = likes_count + 1 WHERE id = $1`,
			commentID)
		liked = true
	}
	if err != nil {
		return false, 0, err
	}

	err = tx.QueryRow(ctx, `SELECT likes_count FROM comments WHERE id = $1`, commentID).Scan(&count)
	if err != nil {
		return false, 0, err
	}

	return liked, count, tx.Commit(ctx)
}
