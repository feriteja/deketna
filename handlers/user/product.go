package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadProduct allows users to view products
func AllProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"products": "products"})
}
