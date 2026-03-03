package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LalitaAng/ecommerce-microservices-grpc/database"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/app"
	"github.com/LalitaAng/ecommerce-microservices-grpc/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	if err := validateEnv(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	dbPool, err := database.NewPool()
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer dbPool.Close()

	container := app.NewContainer(dbPool, jwtSecret)
	defer container.Close()

	r := gin.Default()
	router.SetupRoutes(r, container)
	log.Println("Server running on :8080")
	r.Run(":8080")
}

func validateEnv() error {
	required := []string{"DATABASE_URL", "JWT_SECRET"}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("environment variable %s is required", env)
		}
	}

	if len(os.Getenv("JWT_SECRET")) < 32 {
		return fmt.Errorf("JWT_SECRET is too short")
	}

	return nil
}