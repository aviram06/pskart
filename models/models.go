package models

// Order represents an order in the system
type Order struct {
	OrderId     string  `json:"order_id"`
	UserId      string  `json:"user_id"`
	ItemIds     []int   `json:"item_ids"`
	TotalAmount float64 `json:"total_amount"`
}
