package app

import (
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	DB *pgxpool.Pool

	JWTSecret string

	UserRepo    *user.UserRepository

	UserService    *user.UserService

	UserHandler    *user.UserHandler
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
}

func (c *Container) initServices() {
	c.UserService = user.NewUserService(c.UserRepo, c.JWTSecret)
}

func (c *Container) initHandlers() {
	c.UserHandler = user.NewUserHandler(c.UserService)
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
