package routes

import (
	"tenant-management-go/internal/config"
	"tenant-management-go/internal/handlers"
	"tenant-management-go/internal/keycloak"
	"tenant-management-go/internal/notification"
	"tenant-management-go/internal/repository"
	"tenant-management-go/internal/service"

	"github.com/gin-gonic/gin"
	"database/sql"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Create Keycloak client
	keycloakClient := keycloak.NewKeycloakClient(cfg.Keycloak.BaseURL, cfg.Keycloak.AdminUser, cfg.Keycloak.AdminPass, &cfg.Keycloak)

	// Create notification client
	notificationClient := notification.NewClient(cfg.Notification.BaseURL)

	// Create router group without context path
	api := r.Group("")

	// Account (4 APIs)
	tenantRepo := repository.NewTenantRepository(db)
	tenantService := service.NewTenantService(tenantRepo, keycloakClient, notificationClient)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	api.POST("/account", tenantHandler.CreateTenant)                    // Create
	api.GET("/account", tenantHandler.ListTenants)                      // Search
	api.PUT("/account/:id", tenantHandler.UpdateTenant)                 // Update
	api.DELETE("/account", tenantHandler.DeleteAccount)                 // Delete Account (realm + tenant + config)

	// AccountConfig (3 APIs)
	tenantConfigRepo := repository.NewTenantConfigRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	tenantConfigService := service.NewTenantConfigService(tenantConfigRepo, tenantService, documentRepo)
	tenantConfigHandler := handlers.NewTenantConfigHandler(tenantConfigService, tenantService)
	api.POST("/account/config", tenantConfigHandler.CreateTenantConfig)     // Create
	api.GET("/account/config", tenantConfigHandler.ListTenantConfigs)       // Search
	api.PUT("/account/config/:id", tenantConfigHandler.UpdateTenantConfig)  // Update
}
