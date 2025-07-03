package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
	"localisationgo/pkg/dtos"
)

// MessageHandler is the handler for localization message related APIs
type MessageHandler struct {
	service ports.MessageService
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(service ports.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// Helper to get user ID from header
func getUserIDFromHeader(c *gin.Context) string {
	return c.GetHeader("X-User-ID")
}

// SearchMessages handles the messages search API endpoint
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	locale := c.Query("locale")
	module := c.Query("module")
	codes := c.QueryArray("codes")

	var messages []domain.Message
	var err error

	if len(codes) > 0 {
		messages, err = h.service.SearchMessagesByCodes(c.Request.Context(), tenantID, locale, codes)
	} else {
		messages, err = h.service.SearchMessages(c.Request.Context(), tenantID, module, locale)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// CreateMessages handles the creation of multiple new localization messages
func (h *MessageHandler) CreateMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req dtos.CreateMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromHeader(c)

	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	createdMessages, err := h.service.CreateMessages(c.Request.Context(), tenantID, userID, domainMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"messages": createdMessages})
}

// UpdateMessages handles the update of multiple existing localization messages
func (h *MessageHandler) UpdateMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req dtos.UpdateMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromHeader(c)

	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			UUID:    msg.UUID,
			Message: msg.Message,
		}
	}

	updatedMessages, err := h.service.UpdateMessages(c.Request.Context(), tenantID, userID, domainMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": updatedMessages})
}

// UpsertMessages handles the upsert of multiple localization messages
func (h *MessageHandler) UpsertMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req dtos.UpsertMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromHeader(c)

	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			UUID:    msg.UUID,
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	upsertedMessages, err := h.service.UpsertMessages(c.Request.Context(), tenantID, userID, domainMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": upsertedMessages})
}

// DeleteMessages handles the deletion of multiple localization messages by UUID
func (h *MessageHandler) DeleteMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	uuids := c.QueryArray("uuids")
	if len(uuids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one uuid is required in query params"})
		return
	}

	err := h.service.DeleteMessages(c.Request.Context(), tenantID, uuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// FindMissingMessages handles the API endpoint for finding missing localization messages
func (h *MessageHandler) FindMissingMessages(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	module := c.Query("module")

	missingMessages, err := h.service.FindMissingMessages(c.Request.Context(), tenantID, module)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if missingMessages == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, missingMessages)
}

// BustCache handles clearing the cache for a tenant, with optional module and locale
func (h *MessageHandler) BustCache(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	module := c.Query("module")
	locale := c.Query("locale")

	err := h.service.BustCache(c.Request.Context(), tenantID, module, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache bust operation completed successfully", "success": true})
}

// RegisterRoutes registers the API routes on the given router group
func (h *MessageHandler) RegisterRoutes(router *gin.RouterGroup) {
	// v1 routes
	v1 := router.Group("/localization/messages/v1")
	{
		v1.GET("/_search", h.SearchMessages)
		v1.POST("/_create", h.CreateMessages)
		v1.PUT("/_update", h.UpdateMessages)
		v1.PUT("/_upsert", h.UpsertMessages)
		v1.DELETE("/_delete", h.DeleteMessages)
		v1.GET("/_missing", h.FindMissingMessages)
	}

	// Cache bust route
	cacheBustV1 := router.Group("/localization/cache/v1")
	{
		cacheBustV1.DELETE("/_bust", h.BustCache)
	}
}
