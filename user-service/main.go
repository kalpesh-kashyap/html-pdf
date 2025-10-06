package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/kalpesh-kashyap/html-pdf/user-service/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Database config

	if err := database.ConnectDB(); err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	fmt.Println("✅ Database connected successfully")

	// fiber config
	app := fiber.New(fiber.Config{
		AppName:      "HTML to PDF User Service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())

	// Register routes

	port := os.Getenv("PORT")

	if port == "" {
		port = "3001"
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	go func() {
		fmt.Printf("User service is running on port %s\n", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error", err)
			stop()
		}
	}()
	<-serverCtx.Done()
	fmt.Println("\n Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Server exited gracefully")
}
