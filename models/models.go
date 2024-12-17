package models

import (
	"database/sql"
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
	DeletedAt sql.NullTime `gorm:"index"`

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
	ID         uint    `gorm:"primaryKey"`
	Name       string  `gorm:"size:255;not null"`
	Price      float64 `gorm:"not null;check:price > 0"`
	Stock      int     `gorm:"not null;check:stock >= 0"`
	SellerID   uint    `gorm:"not null"`
	CategoryID *uint   // Optional field for category
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Seller   User     `gorm:"foreignKey:SellerID;constraint:OnDelete:CASCADE"`
	Category Category `gorm:"foreignKey:CategoryID"`
	Orders   []Order  `gorm:"foreignKey:ProductID"`
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

	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Buyer   User    `gorm:"foreignKey:BuyerID;constraint:OnDelete:CASCADE"`
}

type OrderItem struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	OrderID   uint64    `json:"order_id"`
	ProductID uint64    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
