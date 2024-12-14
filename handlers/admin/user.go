package admin

import (
	"deketna/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadUsers allows admins to view all users
func ReadUsers(c *gin.Context) {
	rows, err := config.DB.Query(c, "SELECT id, name, email FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			continue
		}
		users = append(users, gin.H{"id": id, "name": name, "email": email})
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
