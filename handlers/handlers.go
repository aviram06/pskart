package handlers

import (
	"database/sql"
	"log"
	"pskart/models"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

var (
	orderQueue = make(map[string]string)
	queueMutex = sync.RWMutex{}
	metrics    = map[string]int{`json:"pending"`: 0, `json:"processing"`: 0, `json:"completed"`: 0}
)

func ProcessOrders(orderChan chan models.Order, db *sql.DB) {
	for order := range orderChan {
		// Update status to Processing
		updateOrderStatus(order.OrderId, "Processing")
		query := `
			INSERT INTO orders (order_id, user_id, item_ids, total_amount)
			VALUES ($1, $2, $3, $4)
		`
		_, err := db.Exec(query, order.OrderId, order.UserId, pq.Array(order.ItemIds), order.TotalAmount)
		if err != nil {
			log.Printf("Error inserting order %s: %v", order.OrderId, err)
			continue
		}

		// Update status to Completed
		updateOrderStatus(order.OrderId, "Completed")
	}
}

// CreateOrder creates a new order and sends it to the order channel
// @Summary Create a new order
// @Description Create a new order and send it to the order channel
// @Tags orders
// @Accept json
// @Produce json
// @Param order body models.Order true "Order"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /orders [post]
func CreateOrder(c *fiber.Ctx, orderChan chan models.Order) error {
	var order models.Order
	if err := c.BodyParser(&order); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	updateOrderStatus(order.OrderId, "Pending")

	orderChan <- order

	return c.Status(202).JSON(fiber.Map{"message": "Order received"})
}

// GetOrderStatus gets the status of an order by its ID
// @Summary Get order status
// @Description Get the status of an order by its ID
// @Tags orders
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /order/{orderId} [get]
func GetOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("orderId")

	queueMutex.RLock()
	status, exists := orderQueue[orderID]
	queueMutex.RUnlock()

	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.Status(200).JSON(fiber.Map{"orderId": orderID, "status": status})
}

// GetMetrics gets the metrics of orders
// @Summary Get order metrics
// @Description Get the metrics of orders
// @Tags orders
// @Produce json
// @Success 200 {object} map[string]int
// @Router /order [get]
func GetMetrics(c *fiber.Ctx) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return c.Status(200).JSON(metrics)
}

func updateOrderStatus(orderID string, status string) {
	queueMutex.RLock()
	defer queueMutex.RUnlock()

	// Update metrics
	if oldStatus, exists := orderQueue[orderID]; exists {
		metrics[oldStatus]--
	}
	metrics[status]++

	orderQueue[orderID] = status
}
