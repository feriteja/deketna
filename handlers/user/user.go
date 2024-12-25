package user

import (
	"net/http"
	"os"
	"time"

	"deketna/config"
	"deketna/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = os.Getenv("JWT_SECRET")

var jwtSecretKey = []byte(jwtSecret)

// CreateUser registers a new user with database integration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags User Auth
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User registration data"
// @Success 200 {object} helper.SuccessResponse{data=SignInResponse} "User Created successfully"
// @Failure 400 {object} helper.ErrorResponse "Bad Request: Invalid input/Email is already registered"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /register [post]
func CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// Bind and validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Invalid input. Ensure email and password are provided correctly."})
		return
	}

	// Check if email already exists
	if isEmailTaken(req.Email) {
		helper.SendError(c, http.StatusInternalServerError, []string{"Email is already registered."})

		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Error hashing password."})
		return
	}

	// Store user in the database
	user := User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "buyer", // Default role
	}
	if err := createUser(&user); err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Error creating user in database."})

		return
	}

	// Generate JWT token
	token, err := generateJWT(user.Email, user.ID, user.Role)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Error generating JWT token."})

		return
	}

	// Return success response
	helper.SendSuccess(c, http.StatusOK, "User Created successfully", gin.H{
		"Token": token,
	})
}

// SignIn authenticates a user and returns a JWT token
// @Summary Sign in a user (buyer)
// @Description Authenticates a user with email and password
// @Tags User Auth
// @Accept json
// @Produce json
// @Param user body SignInRequest true "User sign-in data"
// @Success 200 {object} helper.SuccessResponse{data=SignInResponse} "User Login successfully"
// @Failure 400 {object} helper.ErrorResponse  "Bad Request: Invalid input"
// @Failure 500 {object} helper.ErrorResponse  "Internal Server Error"
// @Router /signin [post]
func SignIn(c *gin.Context) {
	var req SignInRequest

	// Bind and validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Invalid input. Ensure email and password are provided correctly."})
		return
	}

	// Check if the user exists
	user, err := getUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			helper.SendError(c, http.StatusBadRequest, []string{"Invalid email or password."})

		} else {
			helper.SendError(c, http.StatusInternalServerError, []string{"Error checking user credentials."})

		}
		return
	}

	// Compare the provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Invalid email or password."})

		return
	}

	// Generate JWT token
	token, err := generateJWT(user.Email, user.ID, user.Role)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Error generating token."})

		return
	}

	// Return success response with JWT token
	helper.SendSuccess(c, http.StatusOK, "User Login successfully", gin.H{
		"Token": token,
	})
}

// Helper: Get user by email from the database
func getUserByEmail(email string) (User, error) {
	var user User
	result := config.DB.Where("email = ?", email).First(&user)
	return user, result.Error
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
