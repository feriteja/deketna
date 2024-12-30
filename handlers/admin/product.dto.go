package admin

import (
	"mime/multipart"
	"time"
)

type AddProductRequest struct {
	Name       string                `form:"name" binding:"required"`
	Price      float64               `form:"price" binding:"required,gt=0"`
	Stock      int                   `form:"stock" binding:"required,gt=0"`
	CategoryID int                   `form:"category_id" binding:"required, gt=0"`
	Image      *multipart.FileHeader `form:"image" binding:"required"`
}

type GetProductResponse struct {
	ID         uint64    `json:"id" example:"1"`
	Name       string    `json:"name"`
	Price      float64   `json:"price"`
	Stock      int       `json:"stock"`
	SellerID   uint64    `json:"seller_id"`
	CategoryID *uint     `json:"category_id,omitempty"`
	ImageURL   string    `json:"image_url"` // URL or path to the image
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Seller   Profile  `json:"seller"`
	Category Category `json:"category"`
}

type Profile struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type Category struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProductEditRequest struct {
	Name       string   `json:"name"`
	Price      *float64 `json:"price,omitempty"`
	Stock      *int     `json:"stock,omitempty"`
	CategoryID *uint    `json:"category_id,omitempty"`
	ImageURL   *string  `json:"image_url,omitempty"`
}
