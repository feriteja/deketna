package user

type ProductWithSeller struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Stock      int     `json:"stock"`
	ImageURL   string  `json:"image_url"`
	SellerID   uint64  `json:"seller_id"`
	SellerName string  `json:"seller_name,omitempty"` // Omitempty for null values
}

type ProductDetail struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}
