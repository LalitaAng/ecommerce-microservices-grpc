package app

import (
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/user"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/product"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/order"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/payment"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	DB *pgxpool.Pool

	JWTSecret string

	UserRepo    *user.UserRepository
	ProductRepo *product.ProductRepository
	PaymentRepo *payment.PaymentRepository
	OrderRepo   *order.OrderRepository

	UserService    *user.UserService
	ProductService *product.ProductService
	PaymentService *payment.PaymentService
	OrderService   *order.OrderService

	UserHandler    *user.UserHandler
	ProductHandler *product.ProductHandler
	PaymentHandler *payment.PaymentHandler
	OrderHandler   *order.OrderHandler
}

func NewContainer(db *pgxpool.Pool, jwtSecret string) *Container {
	c := &Container{
		DB:        db,
		JWTSecret: jwtSecret,
	}

	c.initRepositories()
	c.initServices()
	c.initHandlers()

	return c
}

func (c *Container) initRepositories() {
	c.UserRepo = user.NewUserRepository(c.DB)
	c.ProductRepo = product.NewProductRepository(c.DB)
	c.PaymentRepo = payment.NewPaymentRepository(c.DB)
	c.OrderRepo = order.NewOrderRepository(c.DB)
}

func (c *Container) initServices() {
	c.UserService = user.NewUserService(c.UserRepo, c.JWTSecret)
	c.ProductService = product.NewProductService(c.ProductRepo)
	c.PaymentService = payment.NewPaymentService(c.PaymentRepo)
	c.OrderService = order.NewOrderService(c.OrderRepo, c.ProductRepo, c.PaymentService)
}

func (c *Container) initHandlers() {
	c.UserHandler = user.NewUserHandler(c.UserService)
	c.ProductHandler = product.NewProductHandler(c.ProductService)
	c.PaymentHandler = payment.NewPaymentHandler(c.PaymentService)
	c.OrderHandler = order.NewOrderHandler(c.OrderService)
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
