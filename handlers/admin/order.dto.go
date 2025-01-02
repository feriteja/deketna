package admin

// OrderBuyerResponse represents buyer details in an order
type OrderBuyerResponse struct {
	ID    uint64 `json:"id" example:"1"`
	Name  string `json:"name" example:"buyer1234"`
	Email string `json:"email" example:"buyer@example.com"`
	Phone string `json:"phone" example:"123456789"`
}

type OrderItemResponse struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	OrderID     uint64  `json:"order_id"`
	TotalPrice  float64 `json:"total_price"`
	ImageURL    string  `json:"image_url"`
}

type OrderResponse struct {
	OrderID     uint64  `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type OrderDetailWithItemsResponse struct {
	OrderResponse
	Buyer OrderBuyerResponse  `json:"buyer"`
	Items []OrderItemResponse `json:"order_items"`
}
