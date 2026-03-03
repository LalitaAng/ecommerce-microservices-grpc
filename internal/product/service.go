package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProductService struct {
	repo *ProductRepository
}

func NewProductService(repo *ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, productRequest CreateProductRequest) (*Product, error) {
	if productRequest.Name == "" {
		return nil, errors.New("product name is required")
	}

	if productRequest.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	if productRequest.StockQuantity < 0 {
		return nil, errors.New("stock quantity must be non-negative")
	}

	product := &Product{
		ID:            uuid.New().String(),
		Name:          productRequest.Name,
		Description:   productRequest.Description,
		Price:         productRequest.Price,
		StockQuantity: productRequest.StockQuantity,
		Category:      productRequest.Category,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %v", err)
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, productRequest UpdateProductRequest) (*Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if productRequest.Name != nil {
		updates["name"] = *productRequest.Name
	}

	if productRequest.Description != nil {
		updates["description"] = *productRequest.Description
	}

	if productRequest.Price != nil {
		updates["price"] = *productRequest.Price
	}

	if productRequest.Category != nil {
		updates["category"] = *productRequest.Category
	}

	if len(updates) == 0 {
		return product, nil
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	updatedProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated product: %v", err)
	}

	return updatedProduct, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, category string, page, pageSize int) ([]*Product, int32, error) {
	if page <= 0 {
		page = PageDefault
	}
	if pageSize <= 0 || pageSize > PageSizeMax {
		pageSize = PageSizeDefault
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, category, pageSize, offset)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
