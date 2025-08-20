package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	localizationv1 "localisationgo/api/proto/localization/v1"
	"localisationgo/configs"
	"localisationgo/internal/cache"
	"localisationgo/internal/core/ports"
	"localisationgo/internal/core/services"
	"localisationgo/internal/handlers"
	"localisationgo/internal/migration"
	dbpostgres "localisationgo/internal/repositories/postgres"
)

func main() {
	// Load application configurations
	config := configs.LoadConfig()

	// Setup database connection
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName))
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}

	// Apply database migrations
	migrationConfig := &migration.Config{
		Enabled: true,
		Path:    "migrations",
		Timeout: 30 * time.Second,
	}
	migrationRunner := migration.NewRunner(db, migrationConfig)
	if err := migrationRunner.Run(context.Background()); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("Database migrations applied successfully")

	// Initialize Cache
	var messageCache ports.MessageCache
	cacheType := os.Getenv("CACHE_TYPE")

	if cacheType == "in-memory" {
		messageCache = cache.NewInMemoryMessageCache()
		log.Println("Initialized in-memory cache.")
	} else {
		// Default to Redis
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
		messageCache = cache.NewRedisMessageCache(redisClient)
		log.Println("Initialized Redis cache.")
	}

	// Initialize repository and service
	messageRepo := dbpostgres.NewMessageRepository(db)
	messageService := services.NewMessageService(messageRepo, messageCache)

	// Load all messages into memory for the missing messages API
	if err := messageService.LoadAllMessages(context.Background()); err != nil {
		log.Fatalf("Failed to load messages into memory map: %v", err)
	}

	// Setup HTTP server
	httpRouter := gin.Default()
	messageHandler := handlers.NewMessageHandler(messageService)
	apiGroup := httpRouter.Group("/localization")
	messageHandler.RegisterRoutes(apiGroup)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RESTPort),
		Handler: httpRouter,
	}

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	localizationGRPCServer := handlers.NewGRPCServer(messageService)
	localizationv1.RegisterLocalizationServiceServer(grpcServer, localizationGRPCServer)

	// Start servers
	go func() {
		log.Printf("HTTP server listening on :%d", config.RESTPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPCPort))
		if err != nil {
			log.Fatalf("gRPC server failed to listen: %v", err)
		}
		log.Printf("gRPC server listening on :%d", config.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}
	grpcServer.GracefulStop()
	log.Println("Servers gracefully stopped")
}
