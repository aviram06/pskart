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
)

// @title PSKart API
// @version 1.0
// @description This is a sample server for PSKart.
// @host localhost:8080
// @BasePath /api
func main() {
	// Connect to database
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

	// CORS configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow all origins
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	orderChan := make(chan models.Order, 100)

	go handlers.ProcessOrders(orderChan, db)

	api := app.Group("/api")
	// Initiating swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	api.Post("/order", func(c *fiber.Ctx) error {
		return handlers.CreateOrder(c, orderChan)
	})
	api.Get("/ordersnv", handlers.GetMetrics)
	api.Get("/order/:orderId", handlers.GetOrderStatus)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Server port not specified")
	}
	log.Fatal(app.Listen(":" + port))
}
