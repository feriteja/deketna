package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	Phone     string `gorm:"size:15"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"type:user_role;not null"` // Referencing the user_role enum type
	CreatedAt time.Time
	UpdatedAt time.Time

	Products     []Product     `gorm:"foreignKey:SellerID"` // One-to-many with Product
	Transactions []Transaction `gorm:"foreignKey:BuyerID"`  // One-to-many with Transaction
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

	Seller       User          `gorm:"foreignKey:SellerID;constraint:OnDelete:CASCADE"`
	Category     Category      `gorm:"foreignKey:CategoryID"`
	Transactions []Transaction `gorm:"foreignKey:ProductID"`
}

type Category struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:255;unique;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Products []Product `gorm:"foreignKey:CategoryID"` // One-to-many with Product
}

type Transaction struct {
	ID         uint    `gorm:"primaryKey"`
	ProductID  uint    `gorm:"not null"`
	BuyerID    uint    `gorm:"not null"`
	Quantity   int     `gorm:"not null;check:quantity > 0"`
	TotalPrice float64 `gorm:"not null;check:total_price > 0"`
	Status     string  `gorm:"type:transaction_status;default:'pending';not null"` // Enum type
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Buyer   User    `gorm:"foreignKey:BuyerID;constraint:OnDelete:CASCADE"`
}

type AuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Action    string `gorm:"type:text;not null"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
