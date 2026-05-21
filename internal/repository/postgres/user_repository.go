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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateInput — yangi user yaratish uchun.
type CreateInput struct {
	Username        string
	Phone           string
	PasswordHash    string
	FullName        string
	Language        string
	PhoneVerifiedAt *time.Time
}

func (r *UserRepository) Create(ctx context.Context, in CreateInput) (*domain.User, error) {
	const query = `
		INSERT INTO users (username, phone, password_hash, full_name, language, phone_verified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, phone, username, full_name, role, is_business, is_verified,
		          language, posts_count, followers_count, following_count,
		          created_at, updated_at, phone_verified_at
	`

	lang := in.Language
	if lang == "" {
		lang = "uz"
	}
	fullName := in.FullName
	if fullName == "" {
		fullName = in.Username
	}

	var u domain.User
	err := r.pool.QueryRow(ctx, query,
		in.Username, in.Phone, in.PasswordHash, fullName, lang, in.PhoneVerifiedAt,
	).Scan(
		&u.ID, &u.Phone, &u.Username, &u.FullName, &u.Role, &u.IsBusiness, &u.IsVerified,
		&u.Language, &u.PostsCount, &u.FollowersCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt, &u.PhoneVerifiedAt,
	)
	if err != nil {
		if isUniqueViolation(err, "users_phone_key") {
			return nil, domain.ErrPhoneAlreadyExists
		}
		if isUniqueViolation(err, "users_username_key") {
			return nil, domain.ErrUsernameAlreadyExists
		}
		return nil, err
	}

	return &u, nil
}

// FindByUsernameOrPhone — login uchun (username yoki telefon).
func (r *UserRepository) FindByUsernameOrPhone(ctx context.Context, identifier string) (*domain.User, string, error) {
	const query = `
		SELECT id, email, phone, password_hash, full_name, username, avatar_url, bio,
		       role, is_business, is_verified, email_verified_at, phone_verified_at,
		       language, country_code, last_active_at,
		       posts_count, followers_count, following_count,
		       created_at, updated_at
		FROM users
		WHERE (username = $1 OR phone = $1) AND deleted_at IS NULL
		LIMIT 1
	`

	var u domain.User
	var passwordHash string
	err := r.pool.QueryRow(ctx, query, identifier).Scan(
		&u.ID, &u.Email, &u.Phone, &passwordHash, &u.FullName, &u.Username,
		&u.AvatarURL, &u.Bio, &u.Role, &u.IsBusiness, &u.IsVerified,
		&u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.Language, &u.CountryCode,
		&u.LastActiveAt, &u.PostsCount, &u.FollowersCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", domain.ErrUserNotFound
		}
		return nil, "", err
	}

	return &u, passwordHash, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, string, error) {
	const query = `
		SELECT id, email, phone, password_hash, full_name, username, avatar_url, bio,
		       role, is_business, is_verified, email_verified_at, phone_verified_at,
		       language, country_code, last_active_at,
		       posts_count, followers_count, following_count,
		       created_at, updated_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`

	var u domain.User
	var passwordHash string
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&u.ID, &u.Email, &u.Phone, &passwordHash, &u.FullName, &u.Username,
		&u.AvatarURL, &u.Bio, &u.Role, &u.IsBusiness, &u.IsVerified,
		&u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.Language, &u.CountryCode,
		&u.LastActiveAt, &u.PostsCount, &u.FollowersCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", domain.ErrUserNotFound
		}
		return nil, "", err
	}

	return &u, passwordHash, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, email, phone, full_name, username, avatar_url, bio,
		       role, is_business, is_verified, email_verified_at, phone_verified_at,
		       language, country_code, last_active_at,
		       posts_count, followers_count, following_count,
		       created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.FullName, &u.Username,
		&u.AvatarURL, &u.Bio, &u.Role, &u.IsBusiness, &u.IsVerified,
		&u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.Language, &u.CountryCode,
		&u.LastActiveAt, &u.PostsCount, &u.FollowersCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) UpdateLastActive(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE users SET last_active_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

type UpdateProfileInput struct {
	FullName  *string
	Username  *string
	Bio       *string
	AvatarURL *string
	Language  *string
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (*domain.User, error) {
	const query = `
		UPDATE users SET
			full_name = COALESCE($2, full_name),
			username = COALESCE($3, username),
			bio = COALESCE($4, bio),
			avatar_url = COALESCE($5, avatar_url),
			language = COALESCE($6, language)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, phone, full_name, username, avatar_url, bio,
		          role, is_business, is_verified, language, country_code,
		          posts_count, followers_count, following_count,
		          created_at, updated_at
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id,
		in.FullName, in.Username, in.Bio, in.AvatarURL, in.Language,
	).Scan(
		&u.ID, &u.Email, &u.Phone, &u.FullName, &u.Username,
		&u.AvatarURL, &u.Bio, &u.Role, &u.IsBusiness, &u.IsVerified,
		&u.Language, &u.CountryCode,
		&u.PostsCount, &u.FollowersCount, &u.FollowingCount,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err, "users_username_key") {
			return nil, errors.New("bu username allaqachon band")
		}
		return nil, err
	}
	return &u, nil
}

// Follow — ikki tomonlama counter yangilanadi.
func (r *UserRepository) Follow(ctx context.Context, follower, target uuid.UUID) (followersCount int, err error) {
	if follower == target {
		return 0, errors.New("o'zingizni kuzata olmaysiz")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, `
		INSERT INTO follows (follower_id, following_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, follower, target)
	if err != nil {
		return 0, err
	}

	if cmd.RowsAffected() > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE users SET followers_count = followers_count + 1 WHERE id = $1`, target)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(ctx,
			`UPDATE users SET following_count = following_count + 1 WHERE id = $1`, follower)
		if err != nil {
			return 0, err
		}
	}

	err = tx.QueryRow(ctx, `SELECT followers_count FROM users WHERE id = $1`, target).Scan(&followersCount)
	if err != nil {
		return 0, err
	}

	return followersCount, tx.Commit(ctx)
}

func (r *UserRepository) Unfollow(ctx context.Context, follower, target uuid.UUID) (followersCount int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		follower, target)
	if err != nil {
		return 0, err
	}

	if cmd.RowsAffected() > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE users SET followers_count = GREATEST(followers_count - 1, 0) WHERE id = $1`, target)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(ctx,
			`UPDATE users SET following_count = GREATEST(following_count - 1, 0) WHERE id = $1`, follower)
		if err != nil {
			return 0, err
		}
	}

	err = tx.QueryRow(ctx, `SELECT followers_count FROM users WHERE id = $1`, target).Scan(&followersCount)
	if err != nil {
		return 0, err
	}

	return followersCount, tx.Commit(ctx)
}

func (r *UserRepository) IsFollowing(ctx context.Context, follower, target uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		follower, target).Scan(&exists)
	return exists, err
}

// Followers — userID'ni kuzatayotganlar ro'yxati.
// viewerID berilsa, har biriga "men kuzatyapmanmi" qo'shiladi.
func (r *UserRepository) Followers(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, limit, offset int) ([]*domain.UserBrief, error) {
	const query = `
		SELECT u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business,
		       CASE WHEN $2::uuid IS NULL THEN FALSE
		            ELSE EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = $2 AND f2.following_id = u.id)
		       END AS following
		FROM follows f
		JOIN users u ON u.id = f.follower_id AND u.deleted_at IS NULL
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4
	`
	return r.scanBriefs(ctx, query, userID, viewerID, limit, offset)
}

// Following — userID kuzatayotgan foydalanuvchilar.
func (r *UserRepository) Following(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID, limit, offset int) ([]*domain.UserBrief, error) {
	const query = `
		SELECT u.id, u.username, u.full_name, u.avatar_url, u.is_verified, u.is_business,
		       CASE WHEN $2::uuid IS NULL THEN FALSE
		            ELSE EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = $2 AND f2.following_id = u.id)
		       END AS following
		FROM follows f
		JOIN users u ON u.id = f.following_id AND u.deleted_at IS NULL
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4
	`
	return r.scanBriefs(ctx, query, userID, viewerID, limit, offset)
}

func (r *UserRepository) scanBriefs(ctx context.Context, query string, args ...any) ([]*domain.UserBrief, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.UserBrief
	for rows.Next() {
		b := &domain.UserBrief{}
		if err := rows.Scan(&b.ID, &b.Username, &b.FullName, &b.AvatarURL,
			&b.IsVerified, &b.IsBusiness, &b.Following); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}
