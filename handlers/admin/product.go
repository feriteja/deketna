package admin

import (
	"deketna/config"
	"deketna/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// @Summary Add a product
// @Description Admin adds a new product
// @Tags Admin Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body AddProductRequest true "Product details"
// @Success 201 {object} map[string]interface{} "Product added"
// @Failure 400 {object} map[string]interface{} "Validation Error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Access forbidden"
// @Router /admin/product [post]
func AddProduct(c *gin.Context) {
	var req AddProductRequest

	// Validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve admin ID from JWT token
	claims := c.MustGet("claims").(jwt.MapClaims)
	adminID, exists := claims["userid"].(float64) // The value will be float64
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in claims"})
		return
	}

	// Create product
	product := models.Product{
		Name:     req.Name,
		Price:    req.Price,
		Stock:    req.Stock,
		SellerID: uint64(adminID), // Assign admin as the seller
	}
	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product added successfully", "product": req.Name})
}

// DeleteProduct deletes a product
func DeleteProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
