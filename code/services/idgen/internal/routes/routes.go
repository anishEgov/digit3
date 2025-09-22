package routes

import (
	"idgen/internal/config"
	"idgen/internal/handlers"
	"idgen/internal/repository"
	"idgen/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Initialize dependencies
	repo := repository.NewIDGenRepository(db)
	svc := service.NewIDGenService(repo)
	handler := handlers.NewIDGenHandler(svc)

	// API routes
	api := router.Group(cfg.ServerContextPath)
	{
		api.POST("/template", handler.RegisterTemplate)
		api.POST("/generate", handler.GenerateID)
	}

	return router
}
