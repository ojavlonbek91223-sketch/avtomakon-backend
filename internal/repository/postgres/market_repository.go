package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avtomakon/backend/internal/domain"
)

type MarketRepository struct {
	pool *pgxpool.Pool
}

func NewMarketRepository(pool *pgxpool.Pool) *MarketRepository {
	return &MarketRepository{pool: pool}
}

// ----- Categories -----

func (r *MarketRepository) Categories(ctx context.Context) ([]*domain.Category, error) {
	const query = `
		SELECT id, parent_id, name, slug, icon_url, order_index
		FROM categories
		ORDER BY order_index ASC, slug ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Category
	for rows.Next() {
		c := &domain.Category{}
		err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.IconURL, &c.OrderIndex)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// ----- Products -----

func (r *MarketRepository) ListProducts(ctx context.Context, p domain.ProductsParams) ([]*domain.Product, error) {
	args := []any{}
	where := "p.is_active = TRUE"

	if p.CategoryID != nil {
		args = append(args, *p.CategoryID)
		where += " AND p.category_id = $" + itoa(len(args))
	}
	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		where += " AND p.name ILIKE $" + itoa(len(args))
	}

	order := "p.sales_count DESC, p.created_at DESC"
	switch p.Sort {
	case "newest":
		order = "p.created_at DESC"
	case "price_asc":
		order = "p.price ASC"
	case "price_desc":
		order = "p.price DESC"
	case "rating":
		order = "p.rating_avg DESC, p.rating_count DESC"
	}

	args = append(args, p.Limit, p.Offset)
	limOff := "LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	query := `
		SELECT p.id, p.name, p.slug, p.brand, p.price, p.original_price,
		       CASE
		         WHEN p.original_price IS NOT NULL AND p.original_price > p.price
		         THEN ROUND(100 - (p.price * 100.0 / p.original_price))::int
		         ELSE 0
		       END AS discount_percent,
		       p.currency, p.stock_quantity, p.rating_avg, p.rating_count,
		       p.is_featured, p.category_id, p.created_at,
		       (SELECT url FROM product_images WHERE product_id = p.id
		        ORDER BY order_index LIMIT 1) AS image_url,
		       b.id, b.name, b.slug
		FROM products p
		LEFT JOIN businesses b ON b.id = p.seller_id
		WHERE ` + where + `
		ORDER BY ` + order + `, p.id ` + `
		` + limOff

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Product
	for rows.Next() {
		p := &domain.Product{Seller: &domain.ProductSeller{}}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Brand, &p.Price, &p.OriginalPrice,
			&p.DiscountPercent, &p.Currency, &p.StockQuantity,
			&p.RatingAvg, &p.RatingCount, &p.IsFeatured, &p.CategoryID,
			&p.CreatedAt, &p.ImageURL,
			&p.Seller.ID, &p.Seller.Name, &p.Seller.Slug,
		)
		if err != nil {
			return nil, err
		}
		p.InStock = p.StockQuantity > 0
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *MarketRepository) FindProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	const query = `
		SELECT p.id, p.name, p.slug, p.brand, p.description, p.price, p.original_price,
		       CASE
		         WHEN p.original_price IS NOT NULL AND p.original_price > p.price
		         THEN ROUND(100 - (p.price * 100.0 / p.original_price))::int
		         ELSE 0
		       END,
		       p.currency, p.stock_quantity, p.rating_avg, p.rating_count,
		       p.is_featured, p.category_id, p.created_at,
		       b.id, b.name, b.slug,
		       COALESCE(array_agg(pi.url ORDER BY pi.order_index)
		                FILTER (WHERE pi.url IS NOT NULL), '{}')
		FROM products p
		LEFT JOIN businesses b ON b.id = p.seller_id
		LEFT JOIN product_images pi ON pi.product_id = p.id
		WHERE p.id = $1 AND p.is_active = TRUE
		GROUP BY p.id, b.id
	`

	p := &domain.Product{Seller: &domain.ProductSeller{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Brand, &p.Description, &p.Price, &p.OriginalPrice,
		&p.DiscountPercent, &p.Currency, &p.StockQuantity,
		&p.RatingAvg, &p.RatingCount, &p.IsFeatured, &p.CategoryID, &p.CreatedAt,
		&p.Seller.ID, &p.Seller.Name, &p.Seller.Slug,
		&p.Images,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("mahsulot topilmadi")
	}
	if err != nil {
		return nil, err
	}
	p.InStock = p.StockQuantity > 0
	if len(p.Images) > 0 {
		p.ImageURL = &p.Images[0]
	}
	return p, nil
}

func (r *MarketRepository) Featured(ctx context.Context, limit int) ([]*domain.Product, error) {
	return r.ListProducts(ctx, domain.ProductsParams{
		Limit: limit,
		Sort:  "popular",
	})
}

// ----- Promotions -----

func (r *MarketRepository) ActivePromotions(ctx context.Context) ([]*domain.Promotion, error) {
	const query = `
		SELECT id, title, subtitle, image_url, link_type, link_target
		FROM promotions
		WHERE is_active = TRUE AND starts_at <= NOW() AND ends_at >= NOW()
		ORDER BY order_index ASC, starts_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Promotion
	for rows.Next() {
		p := &domain.Promotion{}
		err := rows.Scan(&p.ID, &p.Title, &p.Subtitle, &p.ImageURL, &p.LinkType, &p.LinkTarget)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
