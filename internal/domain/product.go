package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID         uuid.UUID         `json:"id"`
	ParentID   *uuid.UUID        `json:"parent_id,omitempty"`
	Name       map[string]string `json:"name"`
	Slug       string            `json:"slug"`
	IconURL    *string           `json:"icon_url,omitempty"`
	OrderIndex int               `json:"order_index"`
}

type Product struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Slug            string            `json:"slug"`
	Brand           *string           `json:"brand,omitempty"`
	Description     *string           `json:"description,omitempty"`
	Price           float64           `json:"price"`
	OriginalPrice   *float64          `json:"original_price,omitempty"`
	DiscountPercent int               `json:"discount_percent"`
	Currency        string            `json:"currency"`
	StockQuantity   int               `json:"stock_quantity"`
	InStock         bool              `json:"in_stock"`
	RatingAvg       float64           `json:"rating_avg"`
	RatingCount     int               `json:"rating_count"`
	IsFeatured      bool              `json:"is_featured"`
	ImageURL        *string           `json:"image_url,omitempty"`
	Images          []string          `json:"images,omitempty"`
	CategoryID      uuid.UUID         `json:"category_id"`
	Seller          *ProductSeller    `json:"seller,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type ProductSeller struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type Promotion struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Subtitle   *string   `json:"subtitle,omitempty"`
	ImageURL   *string   `json:"image_url,omitempty"`
	LinkType   string    `json:"link_type"`
	LinkTarget *string   `json:"link_target,omitempty"`
}

type ProductsParams struct {
	CategoryID *uuid.UUID
	Search     string
	Sort       string // popular, newest, price_asc, price_desc, rating
	Limit      int
	Offset     int
}
