package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo *PaymentRepository
}

func NewPaymentService(repo *PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) CreatePayment(ctx context.Context, paymentRequest CreatePaymentRequest) (*Payment, error) {
	if paymentRequest.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	if paymentRequest.PaymentMethod == "" {
		return nil, errors.New("payment method must be provided")
	}

	payment := &Payment{
		ID:            uuid.New().String(),
		OrderID:       paymentRequest.OrderID,
		Amount:        paymentRequest.Amount,
		Status:        PaymentStatusCompleted,
		PaymentMethod: paymentRequest.PaymentMethod,
		TransactionID: paymentRequest.TransactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("error creating payment: %v", err)
	}

	return payment, nil
}

func (s *PaymentService) ListPayments(ctx context.Context, status string, page, pageSize int) ([]*Payment, int32, error) {
	if page <= 0 {
		page = PageDefault
	}
	if pageSize <= 0 || pageSize > PageSizeMax {
		pageSize = PageSizeDefault
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, status, pageSize, offset)
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PaymentService) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	return s.repo.GetByOrderID(ctx, orderID)
}

func (s *PaymentService) RefundPayment(ctx context.Context, id string) (*Payment, error) {
	return s.repo.RefundPayment(ctx, id)
}

func (s *PaymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error) {
    return s.repo.GetByOrderID(ctx, orderID)
}
