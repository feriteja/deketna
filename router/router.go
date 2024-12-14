package router

import (
	"deketna/handlers/admin"
	"deketna/handlers/user"

	"github.com/gin-gonic/gin"
)

// InitializeRoutes sets up routes for the application
func InitializeRoutes(r *gin.Engine) {
	// Group routes for users
	userRoutes := r.Group("/user")
	{
		userRoutes.POST("/product", user.CreateUser) // Create user
		// userRoutes.PUT("/product", user.UpdateUser)    // Update user
		// userRoutes.GET("/product", user.DetailProduct) // Read product
		// userRoutes.GET("/products", user.ManyProduct)  // Read product
	}

	// Group routes for admins
	adminRoutes := r.Group("/admin")
	{
		// adminRoutes.POST("/user", admin.CreateUser)   // Create user
		adminRoutes.GET("/user", admin.ReadUsers) // Read all users
		// adminRoutes.PUT("/user", admin.UpdateUser)    // Update user
		// adminRoutes.DELETE("/user", admin.DeleteUser) // Delete user

		adminRoutes.POST("/product", admin.CreateProduct) // Create product
		// adminRoutes.GET("/product", admin.DetailProducts)   // Read products
		// adminRoutes.GET("/products", admin.ManyProducts)    // Read products
		// adminRoutes.PUT("/product", admin.UpdateProduct)    // Update product
		adminRoutes.DELETE("/product", admin.DeleteProduct) // Delete product
	}
}
