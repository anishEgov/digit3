package routes

import (
	"notification/internal/config"
	"notification/internal/email"
	"notification/internal/handlers"
	"notification/internal/repository"
	"notification/internal/service"
	"notification/internal/sms"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) (*gin.Engine, *service.EmailService, *service.SMSService) {
	router := gin.Default()

	// Initialize dependencies
	templateRepo := repository.NewTemplateRepository(db)
	templateService := service.NewTemplateService(templateRepo, cfg)
	templateHandler := handlers.NewTemplateHandler(templateService)

	emailProvider := email.NewGmailProvider(cfg)
	emailService := service.NewEmailService(templateService, emailProvider, cfg)
	smsProvider := sms.NewSMSCountryProvider(cfg)
	smsService := service.NewSMSService(templateService, smsProvider, cfg)
	notificationHandler := handlers.NewNotificationHandler(emailService, smsService)

	// API routes
	api := router.Group(cfg.ServerContextPath)
	{
		// Notification Template Management Routes
		template := api.Group("/template")
		{
			template.POST("", templateHandler.CreateTemplate)
			template.PUT("", templateHandler.UpdateTemplate)
			template.GET("", templateHandler.SearchTemplates)
			template.DELETE("", templateHandler.DeleteTemplate)
			template.POST("/preview", templateHandler.PreviewTemplate)
		}

		api.POST("/email/send", notificationHandler.SendEmail)
		api.POST("/sms/send", notificationHandler.SendSMS)
	}

	return router, emailService, smsService
}
