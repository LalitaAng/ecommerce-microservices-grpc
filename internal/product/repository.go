package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (id, name, description, price, stock_quantity, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.StockQuantity,
		product.Category,
		product.CreatedAt,
		product.UpdatedAt,
	)

	return err
}

func (r *ProductRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	setParts := []string{}
	args := []interface{}{}
	argPos := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argPos))
		args = append(args, value)
		argPos += 1
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE products SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(setParts, ", "), argPos)
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*Product, error) {
	var product Product
	query := `
		SELECT id, name, description, price, stock_quantity, category, created_at, updated_at 
		FROM products
		WHERE id = $1
	`

	if err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.StockQuantity,
		&product.Category,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, category string, limit, offset int) ([]*Product, int32, error) {
	var query string
	var countQuery string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, name, description, price, stock_quantity, category, created_at, updated_at 
			FROM products
			WHERE category = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`

		countQuery = `SELECT COUNT(*) FROM products WHERE category = $1`
		args = []interface{}{category, limit, offset}
	} else {
		query = `
			SELECT id, name, description, price, stock_quantity, category, created_at, updated_at 
			FROM products
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`

		countQuery = `SELECT COUNT(*) FROM products`
		args = []interface{}{limit, offset}
	}

	var totalCount int32
	if category != "" {
		r.db.QueryRow(ctx, countQuery, category).Scan(&totalCount)
	} else {
		r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	}

	// get products
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.StockQuantity,
			&product.Category,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		products = append(products, &product)
	}

	return products, totalCount, rows.Err()
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *ProductRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id string) (*Product, error) {
	var product Product
	query := `
		SELECT id, name, description, price, stock_quantity, category, created_at, updated_at 
		FROM products
		WHERE id = $1
		FOR UPDATE
	`

	if err := tx.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.StockQuantity,
		&product.Category,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) DecrementProductStock(ctx context.Context, tx pgx.Tx, productID string, quantity int32) error {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity - $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := tx.Exec(ctx, query, quantity, productID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("product not found")
	}

	return nil
}
