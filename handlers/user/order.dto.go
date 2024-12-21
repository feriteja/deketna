package user

type OrderItemResponse struct {
	ProductName string  `json:"product_name" example:"Product A"`
	Quantity    int     `json:"quantity" example:"2"`
	Price       float64 `json:"price" example:"25.00"`
}

// OrderResponse represents a single order
type OrderResponse struct {
	OrderID     uint64              `json:"order_id" example:"1"`
	TotalAmount float64             `json:"total_amount" example:"75.50"`
	Status      string              `json:"status" example:"completed"`
	Items       []OrderItemResponse `json:"items"`
}

type PaginatedOrdersResponse struct {
	Page      int             `json:"page" example:"1"`       // Current page number
	Limit     int             `json:"limit" example:"10"`     // Number of items per page
	TotalItem int64           `json:"totalItem" example:"25"` // Total number of items
	IsNext    bool            `json:"isNext" example:"true"`  // Whether there is a next page
	IsPrev    bool            `json:"isPrev" example:"false"` // Whether there is a previous page
	Data      []OrderResponse `json:"data"`                   // List of orders
}
