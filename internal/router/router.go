package router

import (
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/app"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, container *app.Container) {
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", container.UserHandler.Register)
			users.POST("/login", container.UserHandler.Login)

			userProtected := users.Group("")
			userProtected.Use(middleware.AuthMiddleware())
			{
				userProtected.GET("/:user_id", container.UserHandler.GetUser)
			}
		}

		products := v1.Group("/products")
		{
			products.GET("", container.ProductHandler.ListProducts)
			products.GET("/:product_id", container.ProductHandler.GetProduct)

			productProtected := products.Group("")
			productProtected.Use(middleware.AuthMiddleware())
			{
				productProtected.POST("", container.ProductHandler.CreateProduct)
				productProtected.PUT("/:product_id", container.ProductHandler.UpdateProduct)
				productProtected.DELETE("/:product_id", container.ProductHandler.DeleteProduct)
			}
		}
	}
}
