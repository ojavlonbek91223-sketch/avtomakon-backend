package domain

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationStatus string

const (
	StatusPending          ApplicationStatus = "pending"
	StatusApproved         ApplicationStatus = "approved"
	StatusRejected         ApplicationStatus = "rejected"
	StatusRequiresChanges  ApplicationStatus = "requires_changes"
)

type BusinessApplication struct {
	ID              uuid.UUID         `json:"id"`
	UserID          uuid.UUID         `json:"user_id"`
	Type            BusinessType      `json:"type"`
	BusinessName    string            `json:"business_name"`
	ContactPhone    string            `json:"contact_phone"`
	Address         string            `json:"address"`
	Location        *GeoPoint         `json:"location,omitempty"`
	ExperienceYears *int              `json:"experience_years,omitempty"`
	Description     *string           `json:"description,omitempty"`
	WorkplacePhotos []string          `json:"workplace_photos"`
	DocumentURL     *string           `json:"document_url,omitempty"`
	Status          ApplicationStatus `json:"status"`
	AdminNotes      *string           `json:"admin_notes,omitempty"`
	ReviewedAt      *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type ApplyBusinessInput struct {
	Type            BusinessType `json:"type" validate:"required,oneof=master seller"`
	BusinessName    string       `json:"business_name" validate:"required,min=2,max=150"`
	ContactPhone    string       `json:"contact_phone" validate:"required,e164"`
	Address         string       `json:"address" validate:"required,min=5"`
	Location        *GeoPoint    `json:"location" validate:"required"`
	ExperienceYears *int         `json:"experience_years" validate:"omitempty,min=0,max=100"`
	Description     string       `json:"description" validate:"omitempty,max=1000"`
	WorkplacePhotos []string     `json:"workplace_photos" validate:"omitempty,max=10,dive,url"`
	DocumentURL     *string      `json:"document_url" validate:"omitempty,url"`
}

var (
	ErrApplicationAlreadyPending = postError("sizning arizangiz allaqachon ko'rib chiqilmoqda")
)
