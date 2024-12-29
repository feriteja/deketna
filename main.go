package main

import (
	"deketna/config"
	"deketna/middleware"
	"deketna/router"
	"log"
	"os"

	_ "deketna/docs" // Import Swagger docs

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Deketna API
// @version 1.0
// @description API for Deketna business application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @type apiKey
// @in header
// @name Authorization
// @description Enter "Bearer <token>" (e.g., "Bearer abc123") as the value.

func main() {

	// Load environment variables (optional)
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Fatalf("Error loading .env.dev file: %v", err)
	}

	// Connect to database
	config.ConnectDB()

	// Set up router
	r := gin.Default()
	r.Static("/uploads", "./uploads")
	r.Use(middleware.CORSMiddleware())

	r.SetTrustedProxies(nil)
	// Routes
	router.InitializeRoutes(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	r.Run(":8080") // Default port is 8080
}

// LoadEnv loads .env file (optional for local dev)
func LoadEnv() error {
	if _, err := os.Stat(".env"); err == nil {
		return godotenv.Load(".env")
	}
	return nil
}
