package admin

import (
	"deketna/config"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = os.Getenv("JWT_SECRET")

var jwtSecretKey = []byte(jwtSecret)

// SignIn authenticates a user and returns a JWT token
// @Summary Sign in a admin
// @Description Authenticates as admin  with email and password
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param admin body SignInRequest true "Admin sign-in data"
// @Success 200 {object} SignInResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/signin [post]
func SignIn(c *gin.Context) {
	var req SignInRequest

	// Bind and validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input. Ensure email and password are provided correctly."})
		return
	}

	// Check if the user exists
	user, err := getUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password."})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error checking user credentials."})
		}
		return
	}

	// Compare the provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password."})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.Email, user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error generating JWT token."})
		return
	}

	// Return success response with JWT token
	c.JSON(http.StatusOK, SignInResponse{Token: token})
}

// Helper: Get user by email from the database
func getUserByEmail(email string) (User, error) {
	var user User
	result := config.DB.Where(&User{Email: email, Role: "admin"}).First(&user)
	fmt.Print(result)
	return user, result.Error
}

// Helper: Generate JWT
func generateJWT(email string, userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"email":  email,
		"userid": userID,
		"role":   role,
		"exp":    time.Now().Add(72 * time.Hour).Unix(), // Expires in 72 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}
