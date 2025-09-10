package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boundary-go/internal/config"
	"boundary-go/internal/handlers"
	"boundary-go/internal/repository"
	"boundary-go/internal/service"
	"boundary-go/pkg/cache"
	"boundary-go/pkg/postgres"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/gorm"
)

func initDatabase(db *gorm.DB) error {
	// Get underlying sql.DB to execute raw SQL migrations
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying database: %v", err)
	}

	// Read and execute boundary table migration
	boundarySQL, err := os.ReadFile("db/migrations/001_create_boundary_table.sql")
	if err != nil {
		return fmt.Errorf("error reading boundary migration: %v", err)
	}
	if _, err := sqlDB.Exec(string(boundarySQL)); err != nil {
		return fmt.Errorf("error creating boundary table: %v", err)
	}

	// Read and execute boundary hierarchy table migration
	hierarchySQL, err := os.ReadFile("db/migrations/002_create_boundary_hierarchy_table.sql")
	if err != nil {
		return fmt.Errorf("error reading hierarchy migration: %v", err)
	}
	if _, err := sqlDB.Exec(string(hierarchySQL)); err != nil {
		return fmt.Errorf("error creating boundary hierarchy table: %v", err)
	}

	// Read and execute boundary relationship table migration
	relationshipSQL, err := os.ReadFile("db/migrations/003_create_boundary_relationship_table.sql")
	if err != nil {
		return fmt.Errorf("error reading relationship migration: %v", err)
	}
	if _, err := sqlDB.Exec(string(relationshipSQL)); err != nil {
		return fmt.Errorf("error creating boundary relationship table: %v", err)
	}

	return nil
}

// initTracer initializes OpenTelemetry tracing with Jaeger exporter
func initTracer(cfg *config.Config) func() {
	if !cfg.OpenTelemetry.Enabled {
		log.Println("OpenTelemetry tracing is disabled")
		return func() {}
	}

	// Use OTLP HTTP exporter
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(cfg.OpenTelemetry.OTLPEndpoint), // <-- new field
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Printf("Failed to create OTLP exporter: %v", err)
		return func() {}
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.OpenTelemetry.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		log.Printf("Failed to create resource: %v", err)
		return func() {}
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.OpenTelemetry.SamplingRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	log.Printf("OpenTelemetry tracing initialized with service name: %s", cfg.OpenTelemetry.ServiceName)

	// Return cleanup function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}

func main() {
	// Load configuration from environment variables
	cfg := config.LoadConfig()

	// Initialize OpenTelemetry tracing
	cleanup := initTracer(cfg)
	defer cleanup()

	// Initialize database connection
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize database tables
	if err := initDatabase(db); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Initialize cache
	var cacheImpl cache.Cache
	switch cfg.Cache.Type {
	case "redis":
		cacheImpl = cache.NewRedisCache(cfg.Cache.Redis.Addr, cfg.Cache.Redis.Password, cfg.Cache.Redis.DB)
		log.Println("Using Redis cache")
	default:
		cacheImpl = cache.NewInMemoryCache()
		log.Println("Using in-memory cache")
	}

	// Initialize repositories
	boundaryRepo := repository.NewBoundaryRepository(db, cfg)
	hierarchyRepo := repository.NewBoundaryHierarchyRepository(db, cfg)
	relationshipRepo := repository.NewBoundaryRelationshipRepository(db, cfg)

	// Initialize services
	boundaryService := service.NewBoundaryService(boundaryRepo, cacheImpl)
	hierarchyService := service.NewBoundaryHierarchyService(hierarchyRepo, cacheImpl)
	relationshipService := service.NewBoundaryRelationshipService(relationshipRepo, cacheImpl)

	// Initialize handlers
	boundaryHandler := handlers.NewBoundaryHandler(boundaryService, hierarchyService, relationshipService)

	// Setup Gin router
	router := gin.Default()

	// Add OpenTelemetry middleware for HTTP tracing
	if cfg.OpenTelemetry.Enabled {
		router.Use(otelgin.Middleware(cfg.OpenTelemetry.ServiceName))
	}

	// Register routes at root (no context path prefix)
	api := router.Group("/")
	boundaryHandler.RegisterRoutes(api)

	// Start server
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
