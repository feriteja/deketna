package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User created by User Handler"})
}
