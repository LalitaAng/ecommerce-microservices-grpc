package order

import (
	"time"
)

const (
	PageDefault     int = 1
	PageSizeDefault int = 20
	PageSizeMax     int = 100
)

const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

type Order struct {
	ID            string         `json:"id"`
	UserID        string         `json:"user_id"`
	OrderStatus   string         `json:"order_status"`
	Products      []OrderProduct `json:"products"`
	TotalAmount   float64        `json:"total_amount"`
	PaymentMethod string         `json:"payment_method"`
	PaymentStatus string         `json:"payment_status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type OrderProduct struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

type CreateOrderRequest struct {
	Products      []CreateOrderProduct `json:"products"`
	PaymentMethod string               `json:"payment_method" binding:"required"`
}

type CreateOrderProduct struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required,gt=0"`
}
