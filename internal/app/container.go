package app

import (
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/user"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/product"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	DB *pgxpool.Pool

	JWTSecret string

	UserRepo    *user.UserRepository
	ProductRepo *product.ProductRepository

	UserService    *user.UserService
	ProductService *product.ProductService

	UserHandler    *user.UserHandler
	ProductHandler *product.ProductHandler
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
}

func (c *Container) initServices() {
	c.UserService = user.NewUserService(c.UserRepo, c.JWTSecret)
	c.ProductService = product.NewProductService(c.ProductRepo)
}

func (c *Container) initHandlers() {
	c.UserHandler = user.NewUserHandler(c.UserService)
	c.ProductHandler = product.NewProductHandler(c.ProductService)
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
