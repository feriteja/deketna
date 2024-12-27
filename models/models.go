package models

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	Phone     string `gorm:"size:15"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"type:user_role;not null"` // Referencing the user_role enum type
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Use *time.Time for nullable timestamp

	Products []Product `gorm:"foreignKey:SellerID"` // One-to-many with Product
	Orders   []Order   `gorm:"foreignKey:BuyerID"`  // One-to-many with Transaction
}

type Profile struct {
	ID        uint   `gorm:"primaryKey"`
	Address   string `gorm:"type:text"`
	Name      string `gorm:"size:255"`
	UserID    uint   `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"constraint:OnDelete:CASCADE"` // Belongs to User
}

type Product struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	Name       string    `json:"name"`
	Price      float64   `json:"price"`
	Stock      int       `json:"stock"`
	SellerID   uint64    `json:"seller_id"`
	CategoryID *uint     `json:"category_id,omitempty"`
	ImageURL   string    `json:"image_url"` // URL or path to the image
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Seller   User     `gorm:"foreignKey:SellerID;constraint:OnDelete:CASCADE"`
	Category Category `gorm:"foreignKey:CategoryID"`
}

type Category struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:255;unique;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Products []Product `gorm:"foreignKey:CategoryID"` // One-to-many with Product
}

type AuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Action    string `gorm:"type:text;not null"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Cart struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	BuyerID   uint64     `json:"buyer_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
}

type CartItem struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	CartID    uint64    `json:"cart_id"`
	ProductID uint64    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID          uint64      `gorm:"primaryKey" json:"id"`
	BuyerID     uint64      `json:"buyer_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items"`

	Buyer User `gorm:"foreignKey:BuyerID;constraint:OnDelete:CASCADE"`
}

type OrderItem struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	OrderID   uint64    `json:"order_id"`
	ProductID uint64    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"` // Price at the time of purchase
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Order   Order   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}
