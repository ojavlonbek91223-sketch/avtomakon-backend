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

type OTPRepository struct {
	pool *pgxpool.Pool
}

func NewOTPRepository(pool *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{pool: pool}
}

type CreateOTPInput struct {
	Identifier string
	CodeHash   string
	Purpose    domain.OTPPurpose
	ExpiresAt  time.Time
}

func (r *OTPRepository) Create(ctx context.Context, in CreateOTPInput) error {
	const query = `
		INSERT INTO otp_codes (identifier, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, in.Identifier, in.CodeHash, in.Purpose, in.ExpiresAt)
	return err
}

// FindLatestActive — eng oxirgi amaldagi (used emas) OTP.
func (r *OTPRepository) FindLatestActive(ctx context.Context, identifier string, purpose domain.OTPPurpose) (id uuid.UUID, codeHash string, attempts int, expiresAt time.Time, err error) {
	const query = `
		SELECT id, code_hash, attempts, expires_at
		FROM otp_codes
		WHERE identifier = $1 AND purpose = $2 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	err = r.pool.QueryRow(ctx, query, identifier, purpose).Scan(&id, &codeHash, &attempts, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.ErrOTPInvalid
	}
	return
}

func (r *OTPRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *OTPRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE otp_codes SET used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// LastSentAt — rate limit uchun: oxirgi yuborilgan OTP vaqti.
func (r *OTPRepository) LastSentAt(ctx context.Context, identifier string, purpose domain.OTPPurpose) (*time.Time, error) {
	const query = `
		SELECT created_at FROM otp_codes
		WHERE identifier = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var t time.Time
	err := r.pool.QueryRow(ctx, query, identifier, purpose).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
