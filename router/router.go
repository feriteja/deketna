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
		userRoutes.POST("/", user.CreateUser)
		userRoutes.POST("/product", user.CreateUser)
	}

	// Group routes for admins
	adminRoutes := r.Group("/admin")
	{
		adminRoutes.GET("/user", admin.ReadUsers)
		adminRoutes.POST("/product", admin.CreateProduct)
		adminRoutes.DELETE("/product", admin.DeleteProduct)
	}
}
