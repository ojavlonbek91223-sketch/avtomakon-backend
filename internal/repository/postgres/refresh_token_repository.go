package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

type CreateRefreshInput struct {
	UserID     uuid.UUID
	TokenHash  string
	DeviceInfo map[string]any
	IPAddress  *string
	ExpiresAt  time.Time
}

func (r *RefreshTokenRepository) Create(ctx context.Context, in CreateRefreshInput) (uuid.UUID, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, device_info, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query,
		in.UserID, in.TokenHash, in.DeviceInfo, in.IPAddress, in.ExpiresAt,
	).Scan(&id)
	return id, err
}

// FindActive — token_hash bo'yicha amaldagi (revoke qilinmagan, muddati o'tmagan) token.
func (r *RefreshTokenRepository) FindActive(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`

	var rt domain.RefreshToken
	var userID uuid.UUID
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&id, &userID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRefreshTokenInvalid
		}
		return nil, err
	}

	rt.ID = id.String()
	rt.UserID = userID.String()
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
