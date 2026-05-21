package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// CreateFromCart — savatchani buyurtmaga aylantiradi (tranzaksiya).
func (r *OrderRepository) CreateFromCart(ctx context.Context, userID uuid.UUID, in domain.CreateOrderInput) (*domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Savatchani olish
	var cartID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT id FROM carts WHERE user_id = $1`, userID).Scan(&cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCartEmpty
	}
	if err != nil {
		return nil, err
	}

	// Cart items
	rows, err := tx.Query(ctx, `
		SELECT ci.id, ci.product_id, ci.quantity,
		       p.name, p.price, p.seller_id, p.stock_quantity,
		       (SELECT url FROM product_images WHERE product_id = p.id ORDER BY order_index LIMIT 1)
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1 AND p.is_active = TRUE
	`, cartID)
	if err != nil {
		return nil, err
	}

	type cartLine struct {
		ProductID   uuid.UUID
		Quantity    int
		ProductName string
		Price       float64
		SellerID    uuid.UUID
		Stock       int
		Image       *string
	}
	var lines []cartLine
	for rows.Next() {
		var l cartLine
		var ciID uuid.UUID
		err := rows.Scan(&ciID, &l.ProductID, &l.Quantity, &l.ProductName,
			&l.Price, &l.SellerID, &l.Stock, &l.Image)
		if err != nil {
			rows.Close()
			return nil, err
		}
		lines = append(lines, l)
	}
	rows.Close()

	if len(lines) == 0 {
		return nil, domain.ErrCartEmpty
	}

	// Stock tekshirish + subtotal
	var subtotal float64
	for _, l := range lines {
		if l.Stock < l.Quantity {
			return nil, fmt.Errorf("%s — yetarli zaxira yo'q", l.ProductName)
		}
		subtotal += l.Price * float64(l.Quantity)
	}

	// Delivery fee
	deliveryFee := 0.0
	if in.DeliveryMethod == domain.DeliveryCourier {
		deliveryFee = 25000
	} else if in.DeliveryMethod == domain.DeliveryPost {
		deliveryFee = 35000
	}
	total := subtotal + deliveryFee

	// Buyurtma raqami: AM-2026-XXXXXX
	orderNumber := fmt.Sprintf("AM-%d-%06d",
		time.Now().Year(), time.Now().UnixMilli()%1000000)

	addressJSON, _ := json.Marshal(in.DeliveryAddress)

	order := &domain.Order{
		UserID:          userID,
		OrderNumber:     orderNumber,
		Status:          domain.OrderPending,
		Subtotal:        subtotal,
		DeliveryFee:     deliveryFee,
		Total:           total,
		Currency:        "UZS",
		DeliveryAddress: in.DeliveryAddress,
		DeliveryMethod:  in.DeliveryMethod,
		PaymentMethod:   in.PaymentMethod,
		PaymentStatus:   domain.PaymentPending,
		Items:           make([]*domain.OrderItem, 0, len(lines)),
	}
	if in.Note != "" {
		order.Note = &in.Note
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO orders (order_number, user_id, status, subtotal, delivery_fee, total,
		                    currency, delivery_address, delivery_method, payment_method,
		                    payment_status, note)
		VALUES ($1, $2, $3, $4, $5, $6, 'UZS', $7, $8, $9, 'pending', $10)
		RETURNING id, created_at, updated_at
	`, orderNumber, userID, order.Status, subtotal, deliveryFee, total,
		addressJSON, in.DeliveryMethod, in.PaymentMethod, order.Note,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Order items + stock kamaytirish
	for _, l := range lines {
		itemID := uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, seller_id, quantity, price_at_purchase)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, itemID, order.ID, l.ProductID, l.SellerID, l.Quantity, l.Price)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
			UPDATE products SET stock_quantity = stock_quantity - $1,
			                    sales_count = sales_count + $1
			WHERE id = $2
		`, l.Quantity, l.ProductID)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, &domain.OrderItem{
			ID:              itemID,
			ProductID:       l.ProductID,
			ProductName:     l.ProductName,
			ProductImage:    l.Image,
			Quantity:        l.Quantity,
			PriceAtPurchase: l.Price,
			Total:           l.Price * float64(l.Quantity),
		})
	}

	// Savatchani tozalash
	_, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		return nil, err
	}

	return order, tx.Commit(ctx)
}

func (r *OrderRepository) List(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*domain.Order, error) {
	args := []any{userID, limit, offset}
	where := "o.user_id = $1"
	if status != "" {
		args = append([]any{userID, status, limit, offset}, []any{}...)
		where += " AND o.status = $2"
	}

	limitOffset := "LIMIT $2 OFFSET $3"
	if status != "" {
		limitOffset = "LIMIT $3 OFFSET $4"
	}

	query := `
		SELECT o.id, o.order_number, o.user_id, o.status, o.subtotal, o.delivery_fee, o.total,
		       o.currency, o.delivery_address, o.delivery_method, o.payment_method,
		       o.payment_status, o.note, o.created_at, o.updated_at,
		       (SELECT COUNT(*) FROM order_items WHERE order_id = o.id)
		FROM orders o
		WHERE ` + where + `
		ORDER BY o.created_at DESC
		` + limitOffset

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		var addressRaw []byte
		var itemsCount int
		err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.Total,
			&o.Currency, &addressRaw, &o.DeliveryMethod, &o.PaymentMethod,
			&o.PaymentStatus, &o.Note, &o.CreatedAt, &o.UpdatedAt,
			&itemsCount,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(addressRaw, &o.DeliveryAddress)
		o.Items = make([]*domain.OrderItem, 0, itemsCount)
		orders = append(orders, o)
	}

	// Items'larni yuklash (batch)
	if len(orders) > 0 {
		ids := make([]uuid.UUID, len(orders))
		idx := make(map[uuid.UUID]*domain.Order, len(orders))
		for i, o := range orders {
			ids[i] = o.ID
			idx[o.ID] = o
		}

		itemRows, err := r.pool.Query(ctx, `
			SELECT oi.order_id, oi.id, oi.product_id, p.name, oi.quantity, oi.price_at_purchase,
			       (SELECT url FROM product_images WHERE product_id = oi.product_id ORDER BY order_index LIMIT 1)
			FROM order_items oi
			JOIN products p ON p.id = oi.product_id
			WHERE oi.order_id = ANY($1)
			ORDER BY oi.id
		`, ids)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()
		for itemRows.Next() {
			var orderID uuid.UUID
			it := &domain.OrderItem{}
			err := itemRows.Scan(&orderID, &it.ID, &it.ProductID, &it.ProductName,
				&it.Quantity, &it.PriceAtPurchase, &it.ProductImage)
			if err != nil {
				return nil, err
			}
			it.Total = it.PriceAtPurchase * float64(it.Quantity)
			if o, ok := idx[orderID]; ok {
				o.Items = append(o.Items, it)
			}
		}
	}

	return orders, rows.Err()
}

func (r *OrderRepository) Cancel(ctx context.Context, orderID, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Faqat pending/confirmed bekor qilinishi mumkin
	cmd, err := tx.Exec(ctx, `
		UPDATE orders SET status = 'cancelled'
		WHERE id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
	`, orderID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("buyurtma topilmadi yoki bekor qilib bo'lmaydi")
	}

	// Stock'ni qaytarish
	_, err = tx.Exec(ctx, `
		UPDATE products p SET stock_quantity = p.stock_quantity + oi.quantity
		FROM order_items oi WHERE oi.order_id = $1 AND oi.product_id = p.id
	`, orderID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
