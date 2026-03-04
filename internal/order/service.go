package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/payment"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/product"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrderService struct {
	orderRepo   *OrderRepository
	productRepo *product.ProductRepository
	paymentService *payment.PaymentService
}

func NewOrderService(repo *OrderRepository, productRepo *product.ProductRepository, paymentService *payment.PaymentService) *OrderService {
	return &OrderService{
		orderRepo:   repo,
		productRepo: productRepo,
		paymentService: paymentService,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, orderRequest CreateOrderRequest) (*Order, error) {
	if len(orderRequest.Products) == 0 {
		return nil, errors.New("products are required")
	}

	tx, err := s.orderRepo.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orderProducts, totalAmount, err := s.validateAndReserveProductStock(ctx, tx, orderRequest.Products)
	if err != nil {
		return nil, fmt.Errorf("failed to validate products: %w", err)
	}

	order := &Order{
		ID:            uuid.New().String(),
		UserID:        userID,
		OrderStatus:   OrderStatusPending,
		Products:      orderProducts,
		TotalAmount:   totalAmount,
		PaymentMethod: orderRequest.PaymentMethod,
		PaymentStatus: payment.PaymentStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.orderRepo.Create(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (s *OrderService) PayOrder(ctx context.Context, orderID string, userID string) (*Order, error) {
    order, err := s.orderRepo.GetByID(ctx, orderID)
    if err != nil {
        return nil, fmt.Errorf("order not found: %w", err)
    }

    if order.UserID != userID {
        return nil, errors.New("order does not belong to user")
    }

    if order.PaymentStatus == payment.PaymentStatusCompleted {
        return nil, errors.New("order is already paid")
    }
    if order.OrderStatus == OrderStatusCancelled {
        return nil, errors.New("cannot pay for a cancelled order")
    }

    _, err = s.paymentService.CreatePayment(ctx, payment.CreatePaymentRequest{
        OrderID:       order.ID,
        Amount:        order.TotalAmount,
        PaymentMethod: order.PaymentMethod,
        TransactionID: "txn_" + uuid.New().String(),
    })
    if err != nil {
        return nil, fmt.Errorf("payment failed: %w", err)
    }

    if err := s.orderRepo.UpdatePaymentStatus(ctx, order.ID, payment.PaymentStatusCompleted); err != nil {
        return nil, fmt.Errorf("failed to update payment status: %w", err)
    }

    if err := s.orderRepo.UpdateOrderStatus(ctx, order.ID, OrderStatusCompleted); err != nil {
        return nil, fmt.Errorf("failed to update order status: %w", err)
    }

    order.PaymentStatus = payment.PaymentStatusCompleted
    order.OrderStatus = OrderStatusCompleted
    return order, nil
}

func (s *OrderService) validateAndReserveProductStock(ctx context.Context, tx pgx.Tx, requestProducts []CreateOrderProduct) ([]OrderProduct, float64, error) {
	orderProducts := make([]OrderProduct, 0, len(requestProducts))
	totalAmount := 0.0

	for _, requestProduct := range requestProducts {
		productData, err := s.productRepo.GetByIDForUpdate(ctx, tx, requestProduct.ProductID)
		if err != nil {
			return nil, 0, fmt.Errorf("product %s not found: %w", requestProduct.ProductID, err)
		}

		if productData.StockQuantity < requestProduct.Quantity {
			return nil, 0, fmt.Errorf("insufficient stock for product %s: requested %d, available %d", requestProduct.ProductID, requestProduct.Quantity, productData.StockQuantity)
		}

		if err := s.productRepo.DecrementProductStock(ctx, tx, productData.ID, requestProduct.Quantity); err != nil {
			return nil, 0, fmt.Errorf("failed to update product's stock: %w", err)
		}

		subTotal := productData.Price * float64(requestProduct.Quantity)
		orderProduct := OrderProduct{
			ProductID:   productData.ID,
			ProductName: productData.Name,
			Quantity:    int(requestProduct.Quantity),
			UnitPrice:   productData.Price,
			Subtotal:    subTotal,
		}

		orderProducts = append(orderProducts, orderProduct)
		totalAmount += subTotal
	}

	return orderProducts, totalAmount, nil
}

func (s *OrderService) ListOrders(ctx context.Context, status string, page, pageSize int) ([]*Order, int32, error) {
	if page <= 0 {
		page = PageDefault
	}
	if pageSize <= 0 || pageSize > PageSizeMax {
		pageSize = PageSizeDefault
	}

	offset := (page - 1) * pageSize
	return s.orderRepo.List(ctx, status, pageSize, offset)
}

func (s *OrderService) GetOrderDetails(ctx context.Context, orderID string) (*Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string) (*Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return s.orderRepo.Cancel(ctx, order)
}

func (s *OrderService) ListOrderProducts(ctx context.Context, orderID string) ([]OrderProduct, error) {
	return s.orderRepo.GetOrderProducts(ctx, orderID)
}

func (s *OrderService) RefundOrder(ctx context.Context, orderID string, userID string) (*Order, error) {
    order, err := s.orderRepo.GetByID(ctx, orderID)
    if err != nil {
        return nil, fmt.Errorf("order not found: %w", err)
    }

    if order.UserID != userID {
        return nil, errors.New("order does not belong to user")
    }

    if order.PaymentStatus != payment.PaymentStatusCompleted {
        return nil, errors.New("only completed payments can be refunded")
    }

    p, err := s.paymentService.GetPaymentByOrderID(ctx, orderID)
    if err != nil {
        return nil, fmt.Errorf("payment not found: %w", err)
    }

    _, err = s.paymentService.RefundPayment(ctx, p.ID)
    if err != nil {
        return nil, fmt.Errorf("refund failed: %w", err)
    }

    if err := s.orderRepo.UpdatePaymentStatus(ctx, orderID, payment.PaymentStatusRefunded); err != nil {
        return nil, fmt.Errorf("failed to update payment status: %w", err)
    }

    if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, OrderStatusCancelled); err != nil {
        return nil, fmt.Errorf("failed to update order status: %w", err)
    }

    order.PaymentStatus = payment.PaymentStatusRefunded
    order.OrderStatus = OrderStatusCancelled
    return order, nil
}
