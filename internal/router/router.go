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

		payments := v1.Group("/payments")
		payments.Use(middleware.AuthMiddleware())
		{
			payments.GET("", container.PaymentHandler.ListPayments)
			payments.POST("", container.PaymentHandler.CreatePayment)
			payments.GET("/:payment_id", container.PaymentHandler.GetPayment)
			payments.POST("/:payment_id/refund", container.PaymentHandler.RefundPayment)
			payments.GET("/order/:order_id", container.PaymentHandler.GetPaymentForOrder)
		}

		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.GET("", container.OrderHandler.ListOrders)
			orders.POST("", container.OrderHandler.CreateOrder)
			orders.GET("/:order_id", container.OrderHandler.GetOrderDetails)
			orders.POST("/:order_id/pay", container.OrderHandler.PayOrder)
			orders.PUT("/:order_id/cancel", container.OrderHandler.CancelOrder)
			orders.GET("/:order_id/products", container.OrderHandler.ListOrderProducts)
			orders.POST("/:order_id/refund", container.OrderHandler.RefundOrder)
		}
	}
}
