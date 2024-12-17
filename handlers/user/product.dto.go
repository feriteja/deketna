package user

type Product struct {
	ID    uint64  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type ProductDetail struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type AddToCartRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"` // ID of the product
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}
