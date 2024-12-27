package admin

import (
	"deketna/config"
	"deketna/helper"
	"deketna/models"
	"deketna/utils"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// @Summary Add a product
// @Description Admin adds a new product
// @Tags Admin Product
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Product Name"
// @Param price formData number true "Product Price"
// @Param stock formData integer true "Product Stock"
// @Param image formData file true "Product Image"
// @Success 201 {object} helper.SuccessResponse "Product"
// @Success 400 {object} helper.ErrorResponse "Validation Error"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 403 {object} helper.ErrorResponse "Access forbidden"
// @Router /admin/product [post]
func AddProduct(c *gin.Context) {
	// Parse form data
	var req AddProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to upload image"})
		return
	}

	// Retrieve admin ID from JWT token
	claims := c.MustGet("claims").(jwt.MapClaims)
	adminID, exists := claims["userid"].(float64)
	if !exists {
		helper.SendError(c, http.StatusInternalServerError, []string{"User ID not found in claims"})
		return
	}

	// Save file temporarily
	tempFilePath := filepath.Join("tmp", file.Filename)
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{"Failed to save temporary image"})
		return
	}

	fmt.Println("tempFilePath", tempFilePath)
	// Upload to Supabase
	imageURL, err := utils.UploadImageToSupabase(tempFilePath, file.Filename)
	fmt.Print("imageURL", imageURL)
	if err != nil {
		helper.SendError(c, http.StatusBadRequest, []string{fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	// Create product record in DB
	product := models.Product{
		Name:     req.Name,
		Price:    req.Price,
		Stock:    req.Stock,
		SellerID: uint64(adminID),
		ImageURL: imageURL,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		helper.SendError(c, http.StatusInternalServerError, []string{"Failed to add product to database"})
		return
	}

	// Respond with success
	helper.SendSuccess(c, http.StatusCreated, "Product added successfully", product)
}

func DeleteProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
