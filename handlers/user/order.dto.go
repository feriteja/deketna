package user

type OrderItemRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type OrderItemResponse struct {
	OrderID     uint64  `json:"order_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// OrderResponse represents a single order

type OrderResponse struct {
	OrderID     uint64  `json:"order_id" example:"1"`
 	TotalAmount float64 `json:"total_amount" example:"75.50"`
	Status      string  `json:"status" example:"completed"`
	CreatedAt   string  `json:"created_at"` // Changed to string
	UpdatedAt   string  `json:"updated_at"` // Changed to string
}
