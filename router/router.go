package router

import (
	"deketna/handlers/admin"
	"deketna/handlers/user"
	"deketna/middleware"

	"github.com/gin-gonic/gin"
)

// InitializeRoutes sets up routes for the application
func InitializeRoutes(r *gin.Engine) {
	// User Routes
	userRoutes := r.Group("/user")
	{
		userRoutes.POST("/register", user.CreateUser) // User registration
		userRoutes.POST("/signin", user.SignIn)       // User login
		// Example of other user routes (if needed):
		// userRoutes.GET("/products", user.GetProducts)
	}

	// Admin Routes
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.AdminRoleMiddleware()) // Apply admin middleware to all routes in this group
	{
		adminRoutes.POST("/signin", admin.SignIn)               // Admin login
		adminRoutes.POST("/product", admin.AddProduct)          // Add a product
		adminRoutes.DELETE("/product/:id", admin.DeleteProduct) // Delete a product by ID
		// Example of other admin routes (if needed):
		// adminRoutes.PUT("/product/:id", admin.UpdateProduct)
	}
}
