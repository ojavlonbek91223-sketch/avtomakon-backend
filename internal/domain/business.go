package domain

import (
	"time"

	"github.com/google/uuid"
)

type BusinessType string

const (
	BusinessMaster BusinessType = "master"
	BusinessSeller BusinessType = "seller"
)

type Business struct {
	ID             uuid.UUID    `json:"id"`
	Name           string       `json:"name"`
	Slug           string       `json:"slug"`
	Type           BusinessType `json:"type"`
	Description    *string      `json:"description,omitempty"`
	Phone          string       `json:"phone"`
	Address        string       `json:"address"`
	Location       GeoPoint     `json:"location"`
	DistanceMeters *float64     `json:"distance_meters,omitempty"`
	CoverImageURL  *string      `json:"cover_image_url,omitempty"`
	Gallery        []string     `json:"gallery,omitempty"`
	RatingAvg      float64      `json:"rating_avg"`
	RatingCount    int          `json:"rating_count"`
	IsActive       bool         `json:"is_active"`
	Owner          *BusinessOwner `json:"owner,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type BusinessOwner struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
}

// NearbyParams — xarita so'rovi.
type NearbyParams struct {
	Lat     float64
	Lng     float64
	RadiusM int    // metr
	Type    string // master/seller/"" (barchasi)
	Query   string // nom / manzil (shahar, viloyat, tuman) bo'yicha qidiruv
	Limit   int
}

var ErrBusinessNotFound = postError("biznes topilmadi")
