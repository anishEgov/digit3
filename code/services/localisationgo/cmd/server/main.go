package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"localisationgo/configs"
	"localisationgo/internal/core/services"
	"localisationgo/internal/platform/cache"
	"localisationgo/internal/repositories/postgres"
	"localisationgo/internal/server"
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

	// Load all messages into the cache
	if err := messageService.LoadAllMessages(context.Background()); err != nil {
		log.Fatalf("Failed to load messages into cache: %v", err)
	}

	// Create and start the server
	srv := server.NewServer(
		config.RESTPort, // Use RESTPort from config
		config.GRPCPort, // Use GRPCPort from config
		messageService,
	)

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Server started - REST on port %d, gRPC on port %d", config.RESTPort, config.GRPCPort)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown the server
	if err := srv.Stop(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
