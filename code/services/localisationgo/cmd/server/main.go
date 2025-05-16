package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"localisationgo/configs"
	"localisationgo/internal/core/services"
	"localisationgo/internal/handlers"
	"localisationgo/internal/platform/cache"
	"localisationgo/internal/repositories/postgres"
)

func main() {
	// Load configuration
	config := configs.LoadConfig()

	// Setup database connection
	dbConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode,
	)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Execute database schema
	if _, err := db.Exec(postgres.Schema); err != nil {
		log.Fatalf("Failed to create database schema: %v", err)
	}

	// Seed initial data
	seedSQL := postgres.GetSeedDataSQL()
	if _, err := db.Exec(seedSQL); err != nil {
		log.Printf("Note: Failed to seed initial data: %v", err)
		// Don't fail the application if seeding fails
	} else {
		log.Printf("Successfully seeded initial data")
	}

	// Setup Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Initialize repository, cache, and service
	messageRepo := postgres.NewMessageRepository(db)
	messageCache := cache.NewRedisCache(redisClient)
	messageService := services.NewMessageService(messageRepo, messageCache)

	// Initialize handler
	messageHandler := handlers.NewMessageHandler(messageService)

	// Setup Gin router
	router := gin.Default()

	// Add routes
	apiGroup := router.Group("")
	messageHandler.RegisterRoutes(apiGroup)

	// Start server
	log.Printf("Starting server on port %s...", config.ServerPort)
	if err := router.Run(":" + config.ServerPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
