package domain

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID       uuid.UUID    `json:"id"`
	UserID   uuid.UUID    `json:"user_id"`
	Items    []*CartItem  `json:"items"`
	Subtotal float64      `json:"subtotal"`
	Currency string       `json:"currency"`
	Count    int          `json:"count"`
}

type CartItem struct {
	ID        uuid.UUID `json:"id"`
	Product   *Product  `json:"product"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
	AddedAt   time.Time `json:"added_at"`
}

type AddToCartInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1,max=100"`
}

type UpdateCartItemInput struct {
	Quantity int `json:"quantity" validate:"required,min=1,max=100"`
}
