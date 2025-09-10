package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"tenant-management-go/internal/config"
	"tenant-management-go/internal/db"
	"tenant-management-go/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	db.Connect()

	// Set Gin mode from environment (optional)
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	r := gin.Default()

	routes.RegisterRoutes(r, db.DB, cfg)

	port := cfg.Server.Port
	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 