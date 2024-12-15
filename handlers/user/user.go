package user

import (
	"net/http"
	"os"
	"time"

	"deketna/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = os.Getenv("JWT_SECRET")

var jwtSecretKey = []byte("your_jwt_secret_key")

// CreateUser registers a new user with database integration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags User
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User registration data"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user/register [post]
func CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// Bind and validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input. Ensure email and password are provided correctly."})
		return
	}

	// Check if email already exists
	if isEmailTaken(req.Email) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email is already registered."})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error hashing password."})
		return
	}

	// Store user in the database
	user := User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "buyer", // Default role
	}
	if err := createUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error creating user in database."})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.Email, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error generating JWT token."})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, CreateUserResponse{Token: token})
}

// Helper: Check if email already exists
func isEmailTaken(email string) bool {
	var count int64
	config.DB.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// Helper: Store user in the database
func createUser(user *User) error {
	result := config.DB.Create(user)
	return result.Error
}

// Helper: Generate JWT
func generateJWT(email string, userID uint) (string, error) {
	claims := jwt.MapClaims{
		"email":  email,
		"userid": userID,
		"exp":    time.Now().Add(72 * time.Hour).Unix(), // Expires in 72 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// User represents the user table in the database
