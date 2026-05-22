package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type BusinessRepository struct {
	pool *pgxpool.Pool
}

func NewBusinessRepository(pool *pgxpool.Pool) *BusinessRepository {
	return &BusinessRepository{pool: pool}
}

// Nearby — radius ichidagi bizneslarni masofa bo'yicha tartibda qaytaradi.
// PostGIS ST_DWithin (geography) — radius metrda.
func (r *BusinessRepository) Nearby(ctx context.Context, p domain.NearbyParams) ([]*domain.Business, error) {
	// $1 = lng, $2 = lat — masofa hisoblash uchun doim kerak.
	args := []any{p.Lng, p.Lat}
	conds := []string{"b.is_active = TRUE"}

	// Matnli qidiruv bo'lsa radiusni e'tiborsiz qoldiramiz (butun bazadan
	// qidiriladi — foydalanuvchi boshqa shahar/tumandagi joyni ham topa olsin).
	if p.Query == "" {
		args = append(args, p.RadiusM)
		conds = append(conds, fmt.Sprintf(
			"ST_DWithin(b.location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $%d)",
			len(args)))
	}

	if p.Type != "" {
		args = append(args, p.Type)
		conds = append(conds, fmt.Sprintf("b.type = $%d", len(args)))
	}

	if p.Query != "" {
		args = append(args, "%"+p.Query+"%")
		conds = append(conds, fmt.Sprintf(
			"(b.name ILIKE $%d OR b.address ILIKE $%d)", len(args), len(args)))
	}

	args = append(args, p.Limit)
	limitParam := fmt.Sprintf("$%d", len(args))

	query := `
		SELECT b.id, b.name, b.slug, b.type, b.description, b.phone, b.address,
		       ST_Y(b.location::geometry) AS lat,
		       ST_X(b.location::geometry) AS lng,
		       ST_Distance(b.location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) AS meters,
		       b.cover_image_url, b.gallery, b.rating_avg, b.rating_count,
		       b.is_active, b.created_at,
		       u.id, u.full_name, u.avatar_url, u.is_verified
		FROM businesses b
		JOIN users u ON u.id = b.owner_id
		WHERE ` + strings.Join(conds, " AND ") + `
		ORDER BY meters ASC
		LIMIT ` + limitParam

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*domain.Business
	for rows.Next() {
		b := &domain.Business{Owner: &domain.BusinessOwner{}}
		var lat, lng float64
		var meters float64
		err := rows.Scan(
			&b.ID, &b.Name, &b.Slug, &b.Type, &b.Description, &b.Phone, &b.Address,
			&lat, &lng, &meters,
			&b.CoverImageURL, &b.Gallery, &b.RatingAvg, &b.RatingCount,
			&b.IsActive, &b.CreatedAt,
			&b.Owner.ID, &b.Owner.FullName, &b.Owner.AvatarURL, &b.Owner.IsVerified,
		)
		if err != nil {
			return nil, err
		}
		b.Location = domain.GeoPoint{Lat: lat, Lng: lng}
		b.DistanceMeters = &meters
		businesses = append(businesses, b)
	}

	return businesses, rows.Err()
}

func (r *BusinessRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Business, error) {
	const query = `
		SELECT b.id, b.name, b.slug, b.type, b.description, b.phone, b.address,
		       ST_Y(b.location::geometry), ST_X(b.location::geometry),
		       b.cover_image_url, b.gallery, b.rating_avg, b.rating_count,
		       b.is_active, b.created_at,
		       u.id, u.full_name, u.avatar_url, u.is_verified
		FROM businesses b
		JOIN users u ON u.id = b.owner_id
		WHERE b.id = $1
	`

	b := &domain.Business{Owner: &domain.BusinessOwner{}}
	var lat, lng float64
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Name, &b.Slug, &b.Type, &b.Description, &b.Phone, &b.Address,
		&lat, &lng,
		&b.CoverImageURL, &b.Gallery, &b.RatingAvg, &b.RatingCount,
		&b.IsActive, &b.CreatedAt,
		&b.Owner.ID, &b.Owner.FullName, &b.Owner.AvatarURL, &b.Owner.IsVerified,
	)
	if err != nil {
		return nil, domain.ErrBusinessNotFound
	}
	b.Location = domain.GeoPoint{Lat: lat, Lng: lng}
	return b, nil
}
