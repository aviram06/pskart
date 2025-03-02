package main

import (
	"database/sql"
	"log"
	"os"
	_ "pskart/docs"
	"pskart/handlers"
	"pskart/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// @title PSKart API
// @version 1.0
// @description This is a sample server for PSKart.
// @termsOfService http://pskart.com/terms/

// @contact.name API Support
// @contact.url http://www.pskart.com/support
// @contact.email support@pskart.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
func main() {
	// Connect to database
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()
	// Middleware logging
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${method} ${path} - ${status} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	orderChan := make(chan models.Order, 100)

	go handlers.ProcessOrders(orderChan, db)

	api := app.Group("/api")

	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	api.Post("/orders", func(c *fiber.Ctx) error {
		return handlers.CreateOrder(c, orderChan)
	})
	api.Get("/order", handlers.GetMetrics)
	api.Get("/order/:orderId", handlers.GetOrderStatus)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Server port not specified")
	}
	log.Fatal(app.Listen(":" + port))
}
