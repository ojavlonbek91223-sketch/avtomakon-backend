package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderConfirmed OrderStatus = "confirmed"
	OrderPaid      OrderStatus = "paid"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
	OrderRefunded  OrderStatus = "refunded"
)

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentPaid     PaymentStatus = "paid"
	PaymentFailed   PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCard  PaymentMethod = "card"
	PaymentMethodClick PaymentMethod = "click"
	PaymentMethodPayme PaymentMethod = "payme"
	PaymentMethodCash  PaymentMethod = "cash"
)

type DeliveryMethod string

const (
	DeliveryPickup  DeliveryMethod = "pickup"
	DeliveryCourier DeliveryMethod = "courier"
	DeliveryPost    DeliveryMethod = "post"
)

type Order struct {
	ID              uuid.UUID         `json:"id"`
	OrderNumber     string            `json:"order_number"`
	UserID          uuid.UUID         `json:"user_id"`
	Status          OrderStatus       `json:"status"`
	Subtotal        float64           `json:"subtotal"`
	DeliveryFee     float64           `json:"delivery_fee"`
	Total           float64           `json:"total"`
	Currency        string            `json:"currency"`
	DeliveryAddress map[string]any    `json:"delivery_address"`
	DeliveryMethod  DeliveryMethod    `json:"delivery_method"`
	PaymentMethod   PaymentMethod     `json:"payment_method"`
	PaymentStatus   PaymentStatus     `json:"payment_status"`
	Note            *string           `json:"note,omitempty"`
	Items           []*OrderItem      `json:"items"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type OrderItem struct {
	ID              uuid.UUID `json:"id"`
	ProductID       uuid.UUID `json:"product_id"`
	ProductName     string    `json:"product_name"`
	ProductImage    *string   `json:"product_image,omitempty"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase float64   `json:"price_at_purchase"`
	Total           float64   `json:"total"`
}

type CreateOrderInput struct {
	DeliveryMethod  DeliveryMethod    `json:"delivery_method" validate:"required,oneof=pickup courier post"`
	DeliveryAddress map[string]any    `json:"delivery_address" validate:"required"`
	PaymentMethod   PaymentMethod     `json:"payment_method" validate:"required,oneof=card click payme cash"`
	Note            string            `json:"note" validate:"omitempty,max=500"`
}

var (
	ErrCartEmpty    = postError("savatcha bo'sh")
	ErrOrderNotFound = postError("buyurtma topilmadi")
)
