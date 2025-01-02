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

type OrderResponse struct {
	OrderID     uint64  `json:"order_id" example:"1"`
	TotalAmount float64 `json:"total_amount" example:"75.50"`
	Status      string  `json:"status" example:"completed"`
	CreatedAt   string  `json:"created_at"` // Changed to string
	UpdatedAt   string  `json:"updated_at"` // Changed to string
}

type OrderDetailResponse struct {
	OrderID     uint64  `json:"order_id"`
	BuyerName   string  `json:"buyer_name"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type OrderItemDetailResponse struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TotalPrice  float64 `json:"total_price"`
	ImageURL    string  `json:"image_url"`
}

type OrderDetailWithItemsResponse struct {
	OrderDetailResponse
	Items []OrderItemDetailResponse `json:"order_items"`
}
