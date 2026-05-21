package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReviewTargetType string

const (
	ReviewTargetBusiness ReviewTargetType = "business"
	ReviewTargetProduct  ReviewTargetType = "product"
	ReviewTargetOrder    ReviewTargetType = "order"
)

type Review struct {
	ID                uuid.UUID         `json:"id"`
	Author            *PostAuthor       `json:"author"`
	TargetType        ReviewTargetType  `json:"target_type"`
	TargetID          uuid.UUID         `json:"target_id"`
	TargetName        *string           `json:"target_name,omitempty"`
	OrderID           *uuid.UUID        `json:"order_id,omitempty"`
	Rating            int               `json:"rating"`
	Text              *string           `json:"text,omitempty"`
	Images            []string          `json:"images,omitempty"`
	SellerReply       *string           `json:"seller_reply,omitempty"`
	SellerRepliedAt   *time.Time        `json:"seller_replied_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
}

type CreateReviewInput struct {
	Rating  int      `json:"rating" validate:"required,min=1,max=5"`
	Text    string   `json:"text" validate:"omitempty,max=1000"`
	Images  []string `json:"images" validate:"omitempty,max=5,dive,url"`
	OrderID *string  `json:"order_id" validate:"omitempty,uuid"`
}

type ReplyToReviewInput struct {
	Text string `json:"text" validate:"required,min=1,max=500"`
}

var (
	ErrReviewExists    = postError("siz allaqachon sharh yozgansiz")
	ErrCantReviewSelf  = postError("o'zingizning biznesingizga sharh yoza olmaysiz")
)
