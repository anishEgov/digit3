package main

import (
	"fmt"
	"log"
	"template-config/internal/config"
	"template-config/internal/db"
	"template-config/internal/migration"
	"template-config/internal/routes"

	"context"
)

func buildPostgresDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Build DSN
	dsn := buildPostgresDSN(cfg)

	// Setup database
	dbConn, err := db.ConnectDSN(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	migrationConfig := &migration.Config{
		Enabled: cfg.MigrationEnabled,
		Path:    cfg.MigrationScriptPath,
		Timeout: cfg.MigrationTimeout,
	}

	migrationRunner := migration.NewRunner(dbConn, migrationConfig)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MigrationTimeout)
	defer cancel()

	if err := migrationRunner.Run(ctx); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Setup routes
	router := routes.SetupRoutes(dbConn, cfg)

	// Start server
	log.Printf("Starting server on :%s", cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
