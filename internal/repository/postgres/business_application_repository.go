package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type BusinessApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewBusinessApplicationRepository(pool *pgxpool.Pool) *BusinessApplicationRepository {
	return &BusinessApplicationRepository{pool: pool}
}

type CreateApplicationInput struct {
	UserID          uuid.UUID
	Type            domain.BusinessType
	BusinessName    string
	ContactPhone    string
	Address         string
	Lat             float64
	Lng             float64
	ExperienceYears *int
	Description     *string
	WorkplacePhotos []string
	DocumentURL     *string
}

func (r *BusinessApplicationRepository) Create(ctx context.Context, in CreateApplicationInput) (*domain.BusinessApplication, error) {
	const query = `
		INSERT INTO business_applications (
			user_id, type, business_name, contact_phone, address,
			location, experience_years, description, workplace_photos, document_url
		) VALUES (
			$1, $2, $3, $4, $5,
			ST_SetSRID(ST_MakePoint($7, $6), 4326)::geography,
			$8, $9, $10, $11
		)
		RETURNING id, user_id, type, business_name, contact_phone, address,
		          ST_Y(location::geometry), ST_X(location::geometry),
		          experience_years, description, workplace_photos, document_url,
		          status, admin_notes, reviewed_at, created_at
	`

	app := &domain.BusinessApplication{}
	var lat, lng float64
	err := r.pool.QueryRow(ctx, query,
		in.UserID, in.Type, in.BusinessName, in.ContactPhone, in.Address,
		in.Lat, in.Lng,
		in.ExperienceYears, in.Description, in.WorkplacePhotos, in.DocumentURL,
	).Scan(
		&app.ID, &app.UserID, &app.Type, &app.BusinessName,
		&app.ContactPhone, &app.Address,
		&lat, &lng,
		&app.ExperienceYears, &app.Description, &app.WorkplacePhotos, &app.DocumentURL,
		&app.Status, &app.AdminNotes, &app.ReviewedAt, &app.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	app.Location = &domain.GeoPoint{Lat: lat, Lng: lng}
	return app, nil
}

func (r *BusinessApplicationRepository) HasPending(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM business_applications
		 WHERE user_id = $1 AND status IN ('pending', 'requires_changes'))`,
		userID).Scan(&exists)
	return exists, err
}

func (r *BusinessApplicationRepository) FindMine(ctx context.Context, userID uuid.UUID) ([]*domain.BusinessApplication, error) {
	const query = `
		SELECT id, user_id, type, business_name, contact_phone, address,
		       COALESCE(ST_Y(location::geometry), 0),
		       COALESCE(ST_X(location::geometry), 0),
		       experience_years, description, workplace_photos, document_url,
		       status, admin_notes, reviewed_at, created_at
		FROM business_applications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.BusinessApplication
	for rows.Next() {
		app := &domain.BusinessApplication{}
		var lat, lng float64
		err := rows.Scan(
			&app.ID, &app.UserID, &app.Type, &app.BusinessName,
			&app.ContactPhone, &app.Address,
			&lat, &lng,
			&app.ExperienceYears, &app.Description, &app.WorkplacePhotos, &app.DocumentURL,
			&app.Status, &app.AdminNotes, &app.ReviewedAt, &app.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		app.Location = &domain.GeoPoint{Lat: lat, Lng: lng}
		list = append(list, app)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BusinessApplicationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.BusinessApplication, error) {
	const query = `
		SELECT id, user_id, type, business_name, contact_phone, address,
		       COALESCE(ST_Y(location::geometry), 0),
		       COALESCE(ST_X(location::geometry), 0),
		       experience_years, description, workplace_photos, document_url,
		       status, admin_notes, reviewed_at, created_at
		FROM business_applications
		WHERE id = $1
	`
	app := &domain.BusinessApplication{}
	var lat, lng float64
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.UserID, &app.Type, &app.BusinessName,
		&app.ContactPhone, &app.Address,
		&lat, &lng,
		&app.ExperienceYears, &app.Description, &app.WorkplacePhotos, &app.DocumentURL,
		&app.Status, &app.AdminNotes, &app.ReviewedAt, &app.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("ariza topilmadi")
	}
	if err != nil {
		return nil, err
	}
	app.Location = &domain.GeoPoint{Lat: lat, Lng: lng}
	return app, nil
}
