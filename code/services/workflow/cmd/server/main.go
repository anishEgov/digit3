package main

import (
	"fmt"
	"log"

	"digit.org/workflow/api"
	"digit.org/workflow/api/handlers"
	"digit.org/workflow/config"
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
	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set up Gin engine
	gin.SetMode(cfg.GinMode)
	router := gin.Default()

	// Simple health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// Initialize components
	processRepo := postgres.NewProcessRepository(db)
	stateRepo := postgres.NewStateRepository(db)
	actionRepo := postgres.NewActionRepository(db)
	instanceRepo := postgres.NewProcessInstanceRepository(db)

	processSvc := service.NewProcessService(processRepo, stateRepo, actionRepo)
	stateSvc := service.NewStateService(stateRepo)
	actionSvc := service.NewActionService(actionRepo)

	guard := security.NewRBACGuard()
	transitionSvc := service.NewTransitionService(instanceRepo, stateRepo, actionRepo, guard)

	processHandler := handlers.NewProcessHandler(processSvc)
	stateHandler := handlers.NewStateHandler(stateSvc)
	actionHandler := handlers.NewActionHandler(actionSvc)
	transitionHandler := handlers.NewTransitionHandler(transitionSvc)

	// Register routes
	api.RegisterAllRoutes(router, processHandler, stateHandler, actionHandler, transitionHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
