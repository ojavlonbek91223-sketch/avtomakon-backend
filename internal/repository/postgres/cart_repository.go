package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type CartRepository struct {
	pool *pgxpool.Pool
}

func NewCartRepository(pool *pgxpool.Pool) *CartRepository {
	return &CartRepository{pool: pool}
}

// ensureCart — userning savatchasi mavjudligini ta'minlaydi.
func (r *CartRepository) ensureCart(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var cartID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO carts (user_id) VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, userID).Scan(&cartID)
	return cartID, err
}

func (r *CartRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	cart := &domain.Cart{
		ID: cartID,
		UserID: userID,
		Currency: "UZS",
		Items: []*domain.CartItem{},
	}

	const query = `
		SELECT ci.id, ci.quantity, ci.added_at,
		       p.id, p.name, p.slug, p.brand, p.price, p.original_price, p.currency,
		       p.stock_quantity, p.rating_avg, p.rating_count,
		       (SELECT url FROM product_images WHERE product_id = p.id ORDER BY order_index LIMIT 1)
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1 AND p.is_active = TRUE
		ORDER BY ci.added_at DESC
	`

	rows, err := r.pool.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &domain.CartItem{Product: &domain.Product{}}
		err := rows.Scan(
			&item.ID, &item.Quantity, &item.AddedAt,
			&item.Product.ID, &item.Product.Name, &item.Product.Slug,
			&item.Product.Brand, &item.Product.Price, &item.Product.OriginalPrice,
			&item.Product.Currency, &item.Product.StockQuantity,
			&item.Product.RatingAvg, &item.Product.RatingCount,
			&item.Product.ImageURL,
		)
		if err != nil {
			return nil, err
		}
		item.Subtotal = item.Product.Price * float64(item.Quantity)
		item.Product.InStock = item.Product.StockQuantity > 0
		cart.Items = append(cart.Items, item)
		cart.Subtotal += item.Subtotal
		cart.Count += item.Quantity
	}

	return cart, rows.Err()
}

func (r *CartRepository) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return err
	}

	// Stock tekshirish
	var stock int
	var isActive bool
	err = r.pool.QueryRow(ctx,
		`SELECT stock_quantity, is_active FROM products WHERE id = $1`,
		productID).Scan(&stock, &isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("mahsulot topilmadi")
	}
	if err != nil {
		return err
	}
	if !isActive {
		return errors.New("mahsulot mavjud emas")
	}
	if stock < quantity {
		return errors.New("yetarli zaxira yo'q")
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity,
		    added_at = NOW()
	`, cartID, productID, quantity)
	return err
}

func (r *CartRepository) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, quantity int) error {
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx,
		`UPDATE cart_items SET quantity = $3 WHERE id = $1 AND cart_id = $2`,
		itemID, cartID, quantity)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("savatcha elementi topilmadi")
	}
	return nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) error {
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`, itemID, cartID)
	return err
}

func (r *CartRepository) Clear(ctx context.Context, userID uuid.UUID) error {
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}
