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
	userRoutes := r.Group("/")
	{
		userRoutes.POST("/register", user.CreateUser)         // User registration
		userRoutes.POST("/signin", user.SignIn)               // User login
		userRoutes.GET("/products", user.GetProducts)         // Get list of products
		userRoutes.GET("/product/:id", user.GetProductDetail) // Get detail of products

		userRoutes.Use(middleware.BuyerRoleMiddleware())
		{
			userRoutes.POST("/cart/:id", user.AddToCart) // add to cart
		}

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
