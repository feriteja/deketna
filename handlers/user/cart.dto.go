package user

// CartItem represents the structure of a cart item in the response
type CartItemResponse struct {
	ID          uint64  `json:"id"`
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}

type AddToCartRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"` // ID of the product
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type DeleteCartRequest struct {
	CartItemIDs []uint64 `json:"cart_item_ids"`
}

type UpdateCartRequest struct {
	CartItemID uint64 `json:"cart_item_id"`
	Quantity   int    `json:"quantity"`
}
