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
	publicRoutes := r.Group("/")
	{
		publicRoutes.POST("/register", user.CreateUser)         // User registration
		publicRoutes.POST("/signin", user.SignIn)               // User login
		publicRoutes.GET("/products", user.GetProducts)         // Get list of products
		publicRoutes.GET("/product/:id", user.GetProductDetail) // Get product details
	}

	// Authenticated Routes (SignInMiddleware)
	authRoutes := r.Group("/")
	authRoutes.Use(middleware.SignInMiddleware()) // Ensure user is authenticated
	{
		authRoutes.GET("/profile", user.GetUserProfile)
		authRoutes.PUT("/profile", user.EditUserProfile)
	}

	// Buyer Routes (SignInMiddleware + BuyerRoleMiddleware)
	buyerRoutes := r.Group("/")
	buyerRoutes.Use(middleware.BuyerRoleMiddleware()) // User must have buyer role
	{
		buyerRoutes.POST("/cart", user.AddToCart)
		buyerRoutes.GET("/cart", user.GetCarts)
		buyerRoutes.DELETE("/cart", user.DeleteCart)
		buyerRoutes.PUT("/cart", user.UpdateCart)

		buyerRoutes.POST("/order", user.PlaceOrder)
		buyerRoutes.GET("/orders", user.ViewOrders)
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
