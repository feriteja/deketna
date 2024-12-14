package user

import (
	"deketna/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadProduct allows users to view products
func AllProduct(c *gin.Context) {
	// Example logic
	rows, err := config.DB.Query(c, "SELECT id, name, price FROM products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var price float64
		if err := rows.Scan(&id, &name, &price); err != nil {
			continue
		}
		products = append(products, gin.H{"id": id, "name": name, "price": price})
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}
