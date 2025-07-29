package main

import (
	"context"
	"log"

	"digit.org/workflow/api"
	"digit.org/workflow/api/handlers"
	"digit.org/workflow/config"
	"digit.org/workflow/internal/migration"
	"digit.org/workflow/internal/repository/postgres"
	"digit.org/workflow/internal/security"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := postgres.NewDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	migrationConfig := &migration.Config{
		Enabled: cfg.Migration.RunMigrations,
		Path:    cfg.Migration.MigrationPath,
		Timeout: cfg.Migration.Timeout,
	}

	migrationRunner := migration.NewRunner(db, migrationConfig)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Migration.Timeout)
	defer cancel()

	if err := migrationRunner.Run(ctx); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repositories
	processRepo := postgres.NewProcessRepository(db)
	stateRepo := postgres.NewStateRepository(db)
	attributeValidationRepo := postgres.NewAttributeValidationRepository(db)
	actionRepo := postgres.NewActionRepository(db, attributeValidationRepo)
	instanceRepo := postgres.NewProcessInstanceRepository(db)
	parallelRepo := postgres.NewParallelExecutionRepository(db)

	// Initialize attribute guard for simple key-value validation
	guard := security.NewAttributeGuard()

	// Initialize services
	processService := service.NewProcessService(processRepo, stateRepo, actionRepo)
	stateService := service.NewStateService(stateRepo)
	actionService := service.NewActionService(actionRepo)
	transitionService := service.NewTransitionService(instanceRepo, stateRepo, actionRepo, processRepo, parallelRepo, guard)

	// Initialize handlers
	processHandler := handlers.NewProcessHandler(processService)
	stateHandler := handlers.NewStateHandler(stateService)
	actionHandler := handlers.NewActionHandler(actionService, stateService)
	transitionHandler := handlers.NewTransitionHandler(transitionService)

	// Initialize Gin router
	router := gin.Default()

	// Register all routes
	api.RegisterAllRoutes(router, processHandler, stateHandler, actionHandler, transitionHandler)

	// Start server
	serverPort := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
