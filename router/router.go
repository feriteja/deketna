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
			userRoutes.POST("/cart", user.AddToCart)
			userRoutes.GET("/cart", user.GetCarts)
			userRoutes.DELETE("/cart", user.DeleteCart)
			userRoutes.PUT("/cart", user.UpdateCart)

			userRoutes.POST("/order", user.PlaceOrder)
			userRoutes.GET("/orders", user.ViewOrders)
		}

	}

	// Admin Routes
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.AdminRoleMiddleware()) // Apply admin middleware to all routes in this group
	{
		adminRoutes.POST("/signin", admin.SignIn)
		adminRoutes.GET("/products", admin.GetProduct)
		adminRoutes.GET("/product/:id", admin.GetProductDetail)
		adminRoutes.POST("/product", admin.AddProduct)
		adminRoutes.DELETE("/product/:id", admin.AdminDeleteProduct)
		adminRoutes.PUT("/product/:id", admin.AdminEditProduct)
		adminRoutes.GET("/orders", admin.ViewOrders)
		adminRoutes.PUT("/orders/:id/status", admin.UpdateOrderStatus)

		// Example of other admin routes (if needed):
		// adminRoutes.PUT("/product/:id", admin.UpdateProduct)
	}
}
