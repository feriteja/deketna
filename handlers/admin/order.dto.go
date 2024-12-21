package admin

// OrderBuyerResponse represents buyer details in an order
type OrderBuyerResponse struct {
	ID    uint64 `json:"id" example:"1"`
	Email string `json:"email" example:"buyer@example.com"`
	Phone string `json:"phone" example:"123456789"`
}

// PaginatedAdminOrdersResponse represents the paginated response for admin orders
type PaginatedAdminOrdersResponse struct {
	Page      int                  `json:"page" example:"1"`
	Limit     int                  `json:"limit" example:"10"`
	TotalItem int64                `json:"totalItem" example:"50"`
	IsNext    bool                 `json:"isNext" example:"true"`
	IsPrev    bool                 `json:"isPrev" example:"false"`
	Data      []AdminOrderResponse `json:"data"`
}

// AdminOrderResponse represents a single order with buyer details
type AdminOrderResponse struct {
	OrderID     uint64              `json:"order_id" example:"1"`
	TotalAmount float64             `json:"total_amount" example:"75.50"`
	Status      string              `json:"status" example:"completed"`
	Buyer       OrderBuyerResponse  `json:"buyer"`
	Items       []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ProductName string  `json:"product_name" example:"Product A"`
	Quantity    int     `json:"quantity" example:"2"`
	Price       float64 `json:"price" example:"25.00"`
}
