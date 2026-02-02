package main

import (
	"log"
	"os"

	"defab-erp/internal/auth"
	"defab-erp/internal/core/db"
	"defab-erp/internal/product"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. DB
	database := db.Connect()
	defer database.Close()

	// 2. Stores (Data Layer)
	authStore := auth.NewStore(database)
	productStore := product.NewStore(database)

	// 3. Handlers (HTTP Layer)
	authHandler := auth.NewHandler(authStore)
	// productHandler := product.NewHandler(productStore)

	// 4. Fiber
	app := fiber.New()

	// 5. Routes
	app.Post("/api/users", authHandler.Register)

	// Quick test route for products
	app.Get("/api/products/test", func(c *fiber.Ctx) error {
		// Just to prove store works
		return c.SendString("Product Store Connected")
	})

	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}
