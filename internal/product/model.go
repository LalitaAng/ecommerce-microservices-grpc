package product

import "time"

const (
	PageDefault     int = 1
	PageSizeDefault int = 20
	PageSizeMax     int = 100
)

type Product struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int32     `json:"stock_quantity"`
	Category      string    `json:"category"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	StockQuantity int32   `json:"stock_quantity" binding:"required,gte=0"`
	Category      string  `json:"category"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" binding:"required,gte=0"`
	Category    *string  `json:"category"`
}
