package admin

import (
	"deketna/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateProduct creates a new product
func CreateProduct(c *gin.Context) {
	var product struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := config.DB.Exec(c, "INSERT INTO products (name, price) VALUES ($1, $2)", product.Name, product.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product created successfully"})
}

// DeleteProduct deletes a product
func DeleteProduct(c *gin.Context) {
	var input struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := config.DB.Exec(c, "DELETE FROM products WHERE id = $1", input.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
