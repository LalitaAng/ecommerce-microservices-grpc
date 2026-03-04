package payment

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, status, payment_method, transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Status,
		payment.PaymentMethod,
		payment.TransactionID,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	return err
}

func (r *PaymentRepository) List(ctx context.Context, status string, limit, offset int) ([]*Payment, int32, error) {
	var query string
	var countQuery string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, order_id, amount, status, payment_method, transaction_id, created_at, updated_at
			FROM payments
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`

		countQuery = `SELECT COUNT(*) FROM payments WHERE status = $1`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, order_id, amount, status, payment_method, transaction_id, created_at, updated_at
			FROM payments
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`

		countQuery = `SELECT COUNT(*) FROM payments`
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

	var payments []*Payment
	for rows.Next() {
		var payment Payment
		if err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.Amount,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.TransactionID,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, &payment)
	}

	return payments, totalCount, rows.Err()
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*Payment, error) {
	var payment Payment
	query := `
		SELECT id, order_id, amount, status, payment_method, transaction_id, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	if err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.TransactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	var payment Payment
	query := `
		SELECT id, order_id, amount, status, payment_method, transaction_id, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`

	if err := r.db.QueryRow(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.TransactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) RefundPayment(ctx context.Context, id string) (*Payment, error) {
	var payment Payment
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query,
		PaymentStatusRefunded,
		time.Now(),
		id,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, errors.New("payment not found")
	}

	return &payment, nil
}
