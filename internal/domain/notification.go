package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationLike             NotificationType = "like"
	NotificationComment          NotificationType = "comment"
	NotificationFollow           NotificationType = "follow"
	NotificationMessage          NotificationType = "message"
	NotificationOrderUpdate      NotificationType = "order_update"
	NotificationBusinessApproved NotificationType = "business_approved"
	NotificationReview           NotificationType = "review"
)

type Notification struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id"`
	Type       NotificationType  `json:"type"`
	Actor      *PostAuthor       `json:"actor,omitempty"`
	EntityType *string           `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID        `json:"entity_id,omitempty"`
	Title      string            `json:"title"`
	Body       *string           `json:"body,omitempty"`
	Data       map[string]any    `json:"data,omitempty"`
	IsRead     bool              `json:"is_read"`
	CreatedAt  time.Time         `json:"created_at"`
}

type PushToken struct {
	Token    string `json:"token" validate:"required,min=10"`
	Platform string `json:"platform" validate:"required,oneof=ios android web"`
}
