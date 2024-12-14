package main

import (
	"deketna/config"
	"deketna/router"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables (optional)
	if err := LoadEnv(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Connect to database
	config.ConnectDB()
	defer config.DB.Close()

	// Set up router
	r := gin.Default()

	r.SetTrustedProxies(nil)
	// Routes
	router.InitializeRoutes(r)

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
