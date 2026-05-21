package domain

import (
	"time"

	"github.com/google/uuid"
)

type LongVideo struct {
	ID             uuid.UUID    `json:"id"`
	Author         *PostAuthor  `json:"author"`
	Title          string       `json:"title"`
	Description    *string      `json:"description,omitempty"`
	VideoURL       string       `json:"video_url"`
	ThumbnailURL   *string      `json:"thumbnail_url,omitempty"`
	DurationSec    int          `json:"duration_sec"`
	Category       *string      `json:"category,omitempty"`
	ViewsCount     int64        `json:"views_count"`
	ReactionsCount int          `json:"reactions_count"`
	CommentsCount  int          `json:"comments_count"`
	CreatedAt      time.Time    `json:"created_at"`
}

type CreateVideoInput struct {
	Title        string  `json:"title" validate:"required,min=3,max=200"`
	Description  string  `json:"description" validate:"omitempty,max=5000"`
	VideoURL     string  `json:"video_url" validate:"required,url"`
	ThumbnailURL string  `json:"thumbnail_url" validate:"omitempty,url"`
	DurationSec  int     `json:"duration_sec" validate:"required,min=1,max=86400"`
	Category     string  `json:"category" validate:"omitempty,max=50"`
}

var (
	ErrVideoNotFound    = postError("video topilmadi")
	ErrOnlyBusinessVideo = postError("faqat usta yoki sotuvchilar video yuborishi mumkin")
)
