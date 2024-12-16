package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = os.Getenv("JWT_SECRET")
var jwtSecretKey = []byte(jwtSecret)

func AdminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Exclude the signin route from the middleware
		if c.Request.URL.Path == "/admin/signin" {
			c.Next()
			return
		}

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token missing"})
			c.Abort()
			return
		}

		// Parse the token, removing the "Bearer " prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the token signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return jwtSecretKey, nil // Secret key
		})

		// Token validation
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims and verify role
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["role"] == nil || claims["role"] != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden: admin role required"})
			c.Abort()
			return
		}

		// Store claims in the context for later use in the handler
		c.Set("claims", claims)

		// Proceed to the next middleware or handler
		c.Next()
	}
}
