package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"idgen/internal"
	"idgen/internal/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	db, err := internal.InitDB(cfg)
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer db.Close()

	// Set Gin mode to release mode by default
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.POST("/template", internal.RegisterTemplateHandler(db))
	r.POST("/generate", internal.GenerateIdHandler(db))

	port := cfg.Server.RestPort
	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 