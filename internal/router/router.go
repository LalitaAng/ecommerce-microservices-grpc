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
	}
}
