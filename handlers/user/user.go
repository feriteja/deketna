package user

import (
	"errors"
	"net/http"
	"os"
	"time"

	"deketna/config"
	"deketna/helper"
	"deketna/models"

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
	if _isEmailTaken(req.Email) {
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
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "buyer", // Default role
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Create User
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Step 2: Create Empty Profile Linked to User
		profile := models.Profile{
			UserID:   user.ID,
			Name:     "",
			ImageURL: "",
		}

		if err := tx.Create(&profile).Error; err != nil {
			return errors.New("failed to create user profile")
		}

		return nil
	})

	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// Generate JWT token
	token, err := _generateJWT(user.Email, user.ID, user.Role)
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
	user, err := _getUserByEmail(req.Email)
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
	token, err := _generateJWT(user.Email, user.ID, user.Role)
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Error generating token."})

		return
	}

	// Return success response with JWT token
	helper.SendSuccess(c, http.StatusOK, "User Login successfully", gin.H{
		"Token": token,
	})
}

// @Summary Get User Profile
// @Description Retrieve the profile of the currently authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helper.SuccessResponse{data=ProfileResponse} "Profile retrieved successfully"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /profile [get]
func GetUserProfile(c *gin.Context) {
	// Extract user ID from JWT claims
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := uint64(claims["userid"].(float64))

	// Fetch profile with associated user
	var profile models.Profile
	err := config.DB.Preload("User").Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve profile"})
		return
	}

	// Map data to DTO
	response := ProfileResponse{
		ID:        profile.ID,
		Address:   profile.Address,
		Name:      profile.Name,
		UserID:    profile.UserID,
		ImageURL:  profile.ImageURL,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
		User: UserResponse{
			ID:        profile.User.ID,
			Email:     profile.User.Email,
			Phone:     profile.User.Phone,
			Role:      profile.User.Role,
			CreatedAt: profile.User.CreatedAt.Format(time.RFC3339),
			UpdatedAt: profile.User.UpdatedAt.Format(time.RFC3339),
		},
	}

	// Send response
	helper.SendSuccess(c, http.StatusOK, "Profile retrieved successfully", response)
}

// @Summary Edit User Profile
// @Description Update the profile details of the currently authenticated user
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body EditProfileRequest true "Profile data to update"
// @Success 200 {object} helper.SuccessResponse{data=EditProfileResponse} "Profile updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 500 {object} helper.ErrorResponse "Internal Server Error"
// @Router /profile [put]
func EditUserProfile(c *gin.Context) {
	// Extract user ID from JWT claims
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := uint64(claims["userid"].(float64))

	// Parse input
	var req EditProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// Fetch user's profile
	var profile models.Profile
	if err := config.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to retrieve profile"})
		return
	}

	// Update profile fields
	if req.Name != "" {
		profile.Name = req.Name
	}
	if req.Address != "" {
		profile.Address = req.Address
	}
	if req.ImageURL != "" {
		profile.ImageURL = req.ImageURL
	}

	// Save updates
	if err := config.DB.Save(&profile).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to update profile"})
		return
	}

	// Map profile data to DTO
	response := EditProfileResponse{
		ID:        profile.ID,
		Address:   profile.Address,
		Name:      profile.Name,
		UserID:    profile.UserID,
		ImageURL:  profile.ImageURL,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}

	// Send response
	helper.SendSuccess(c, http.StatusOK, "Profile updated successfully", response)
}

func _getUserByEmail(email string) (models.User, error) {
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	return user, result.Error
}

func _isEmailTaken(email string) bool {
	var count int64
	config.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func _generateJWT(email string, userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"email":  email,
		"userid": userID,
		"role":   role,
		"exp":    time.Now().Add(72 * time.Hour).Unix(), // Expires in 72 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}
