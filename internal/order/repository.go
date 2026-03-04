package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, tx pgx.Tx, order *Order) error {
	query := `
		INSERT INTO orders (id, user_id, order_status, total_amount, payment_method, payment_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.OrderStatus,
		order.TotalAmount,
		order.PaymentMethod,
		order.PaymentStatus,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	productQuery := `
		INSERT INTO order_products (id, order_id, product_id, product_name, quantity, unit_price, subtotal)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, product := range order.Products {
		_, err := tx.Exec(ctx, productQuery,
			uuid.New().String(),
			order.ID,
			product.ProductID,
			product.ProductName,
			product.Quantity,
			product.UnitPrice,
			product.Subtotal,
		)

		if err != nil {
			return fmt.Errorf("failed to create order product: %w", err)
		}
	}

	return nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
    query := `UPDATE orders SET order_status = $1, updated_at = NOW() WHERE id = $2`
    result, err := r.db.Exec(ctx, query, status, orderID)
    if err != nil {
        return err
    }
    if result.RowsAffected() == 0 {
        return errors.New("order not found")
    }
    return nil
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, orderID string, status string) error {
    query := `
        UPDATE orders
        SET payment_status = $1, updated_at = NOW()
        WHERE id = $2
    `
    result, err := r.db.Exec(ctx, query, status, orderID)
    if err != nil {
        return err
    }
    if result.RowsAffected() == 0 {
        return errors.New("order not found")
    }
    return nil
}

func (r *OrderRepository) List(ctx context.Context, status string, limit, offset int) ([]*Order, int32, error) {
	var query string
	var countQuery string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, user_id, order_status, total_amount, payment_method, payment_status, created_at, updated_at
			FROM orders
			WHERE order_status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`

		countQuery = `SELECT COUNT(*) FROM orders WHERE order_status = $1`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, user_id, order_status, total_amount, payment_method, payment_status, created_at, updated_at
			FROM orders
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`

		countQuery = `SELECT COUNT(*) FROM orders`
		args = []interface{}{limit, offset}
	}

	var totalCount int32
	if status != "" {
		r.db.QueryRow(ctx, countQuery, status).Scan(&totalCount)
	} else {
		r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*Order
	orderIDs := make([]string, 0)

	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderStatus,
			&order.TotalAmount,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, &order)
		orderIDs = append(orderIDs, order.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return orders, totalCount, nil
	}

	productsMap, err := r.getOrderProductsBatch(ctx, orderIDs)
	if err != nil {
		return nil, 0, err
	}

	for _, order := range orders {
		order.Products = productsMap[order.ID]
	}

	return orders, totalCount, nil
}

func (r *OrderRepository) getOrderProductsBatch(ctx context.Context, orderIDs []string) (map[string][]OrderProduct, error) {
	query := `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, subtotal
		FROM order_products
		WHERE order_id = ANY($1)
		ORDER BY order_id, id
	`

	rows, err := r.db.Query(ctx, query, orderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productsMap := make(map[string][]OrderProduct)
	for rows.Next() {
		var orderProduct OrderProduct
		if err := rows.Scan(
			&orderProduct.ID,
			&orderProduct.OrderID,
			&orderProduct.ProductID,
			&orderProduct.ProductName,
			&orderProduct.Quantity,
			&orderProduct.UnitPrice,
			&orderProduct.Subtotal,
		); err != nil {
			return nil, err
		}

		productsMap[orderProduct.OrderID] = append(productsMap[orderProduct.OrderID], orderProduct)
	}

	return productsMap, rows.Err()
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*Order, error) {
	var order Order
	query := `	
		SELECT id, user_id, order_status, total_amount, payment_method, payment_status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	if err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.OrderStatus,
		&order.TotalAmount,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) Cancel(ctx context.Context, order *Order) (*Order, error) {
	query := `
		UPDATE orders
		SET order_status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, user_id, order_status, total_amount, payment_method, payment_status, created_at, updated_at
	`

	updated := &Order{}
    err := r.db.QueryRow(ctx, query, OrderStatusCancelled, time.Now(), order.ID).Scan(
        &updated.ID,
        &updated.UserID,
        &updated.OrderStatus,
        &updated.TotalAmount,
        &updated.PaymentMethod,
        &updated.PaymentStatus,
        &updated.CreatedAt,
        &updated.UpdatedAt,
    )

	if err != nil {
        if err == pgx.ErrNoRows {
            return nil, errors.New("order not found")
        }
        return nil, err
    }
    return updated, nil
}

func (r *OrderRepository) GetOrderProducts(ctx context.Context, orderID string) ([]OrderProduct, error) {
	query := `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, subtotal
		FROM order_products
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderProducts []OrderProduct
	for rows.Next() {
		var orderProduct OrderProduct
		if err := rows.Scan(
			&orderProduct.ID,
			&orderProduct.OrderID,
			&orderProduct.ProductID,
			&orderProduct.ProductName,
			&orderProduct.Quantity,
			&orderProduct.UnitPrice,
			&orderProduct.Subtotal,
		); err != nil {
			return nil, err
		}
		orderProducts = append(orderProducts, orderProduct)
	}

	return orderProducts, nil
}
