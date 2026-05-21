package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

type CreateNotificationInput struct {
	UserID     uuid.UUID
	Type       domain.NotificationType
	ActorID    *uuid.UUID
	EntityType *string
	EntityID   *uuid.UUID
	Title      string
	Body       *string
	Data       map[string]any
}

func (r *NotificationRepository) Create(ctx context.Context, in CreateNotificationInput) (*domain.Notification, error) {
	const query = `
		INSERT INTO notifications (user_id, type, actor_id, entity_type, entity_id, title, body, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	n := &domain.Notification{
		UserID: in.UserID,
		Type: in.Type,
		EntityType: in.EntityType,
		EntityID: in.EntityID,
		Title: in.Title,
		Body: in.Body,
		Data: in.Data,
	}
	err := r.pool.QueryRow(ctx, query,
		in.UserID, in.Type, in.ActorID, in.EntityType, in.EntityID,
		in.Title, in.Body, in.Data,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *NotificationRepository) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit int) ([]*domain.Notification, error) {
	where := "n.user_id = $1"
	if unreadOnly {
		where += " AND n.is_read = FALSE"
	}

	query := `
		SELECT n.id, n.user_id, n.type, n.entity_type, n.entity_id,
		       n.title, n.body, n.data, n.is_read, n.created_at,
		       u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business
		FROM notifications n
		LEFT JOIN users u ON u.id = n.actor_id
		WHERE ` + where + `
		ORDER BY n.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var actorID *uuid.UUID
		var username, fullName, avatarURL *string
		var isVerified, isBusiness *bool

		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.EntityType, &n.EntityID,
			&n.Title, &n.Body, &n.Data, &n.IsRead, &n.CreatedAt,
			&actorID, &username, &fullName, &avatarURL,
			&isVerified, &isBusiness,
		)
		if err != nil {
			return nil, err
		}

		if actorID != nil && fullName != nil {
			n.Actor = &domain.PostAuthor{
				ID: *actorID,
				Username: username,
				FullName: *fullName,
				AvatarURL: avatarURL,
				IsVerified: isVerified != nil && *isVerified,
				IsBusiness: isBusiness != nil && *isBusiness,
			}
		}

		list = append(list, n)
	}
	return list, rows.Err()
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`,
		userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`,
		userID)
	return err
}

func (r *NotificationRepository) SavePushToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO push_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET user_id = $1, platform = $3
	`, userID, token, platform)
	return err
}
