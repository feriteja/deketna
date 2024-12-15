package admin

import (
	"database/sql"
	"time"
)

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type SignInResponse struct {
	Token string `json:"token" example:"your_jwt_token"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid input"`
}

type User struct {
	ID        uint         `gorm:"primaryKey"`
	Email     string       `gorm:"uniqueIndex;not null"`
	Password  string       `gorm:"not null"`
	Role      string       `gorm:"not null;default:'buyer'"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"`
	DeletedAt sql.NullTime `gorm:"index"`
}
