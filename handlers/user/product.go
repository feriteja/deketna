package user

import (
	"deketna/config"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProducts retrieves a list of products
// @Summary Get Products
// @Description Retrieve a list of products available for users
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {array} Product "List of products"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products [get]
func GetProducts(c *gin.Context) {
	var products []Product // Product is your GORM model struct

	// Fetch all products from the database
	if err := config.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	// Return the products as a JSON response
	c.JSON(http.StatusOK, products)
}

// GetProductDetail retrieves the details of a specific product
// @Summary Get Product Detail
// @Description Retrieve detailed information of a specific product by its ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} ProductDetail "Product details"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /product/{id} [get]
func GetProductDetail(c *gin.Context) {
	// Parse product ID from the URL
	idParam := c.Param("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product Product // Product is your GORM model struct

	// Fetch the product by ID from the database
	if err := config.DB.First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product details"})
		}
		return
	}

	// Return the product details as a JSON response
	c.JSON(http.StatusOK, product)
}
