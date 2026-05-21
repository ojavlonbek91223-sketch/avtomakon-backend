package postgres

import (
	"context"

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
	args := []any{p.Lng, p.Lat, p.RadiusM}
	whereType := ""
	if p.Type != "" {
		args = append(args, p.Type)
		whereType = " AND b.type = $4"
	}
	args = append(args, p.Limit)

	limitParam := "$5"
	if p.Type == "" {
		limitParam = "$4"
	}

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
		WHERE b.is_active = TRUE
		  AND ST_DWithin(b.location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
		  ` + whereType + `
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
