package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateProduct creates a new product
func CreateProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Product created successfully"})
}

// DeleteProduct deletes a product
func DeleteProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
