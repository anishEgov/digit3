package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"localisationgo/internal/common/models"
	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
	"localisationgo/pkg/dtos"
)

// MessageHandler handles HTTP requests for the localization service
type MessageHandler struct {
	service ports.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(service ports.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// UpsertMessages handles the messages upsert API endpoint
func (h *MessageHandler) UpsertMessages(c *gin.Context) {
	var req dtos.UpsertMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if req.TenantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantId is required"})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one message is required"})
		return
	}

	// Check for duplicate messages within the request
	messageKeys := make(map[string]bool)
	for i, msg := range req.Messages {
		// Create a unique key for each message
		key := fmt.Sprintf("%s:%s:%s:%s", req.TenantId, msg.Module, msg.Locale, msg.Code)
		if _, exists := messageKeys[key]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("duplicate message at index %d with code '%s', module '%s', locale '%s'", i, msg.Code, msg.Module, msg.Locale)})
			return
		}
		messageKeys[key] = true
	}

	// Call the service
	messages, err := h.service.UpsertMessages(c.Request.Context(), req.TenantId, req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert domain messages to response format
	responseMessages := make([]dtos.MessageResponse, len(messages))
	for i, msg := range messages {
		responseMessages[i] = dtos.MessageResponse{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	// Return the response
	c.JSON(http.StatusOK, dtos.UpsertMessagesResponse{
		Messages: responseMessages,
	})
}

// SearchMessages handles the messages search API endpoint
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	var req dtos.SearchMessagesRequest

	// Try to bind JSON if present in the body
	jsonBindErr := c.ShouldBindJSON(&req)

	// If JSON binding failed or body is empty, check for query parameters
	if jsonBindErr != nil || (jsonBindErr == nil && req.TenantId == "") {
		// Get parameters from URL query
		tenantID := c.Query("tenantId")
		module := c.Query("module")
		locale := c.Query("locale")
		codeStr := c.Query("codes")

		// Parse comma-separated codes if present
		var codes []string
		if codeStr != "" {
			codes = strings.Split(codeStr, ",")
		}

		// Create RequestInfo with minimal data for empty requests
		if req.RequestInfo.APIId == "" {
			req.RequestInfo = models.RequestInfo{
				APIId: "api.localization",
				Ver:   "1.0",
				Ts:    time.Now(),
				MsgId: uuid.New().String(),
			}
		}

		// Update request with query parameters if provided
		if tenantID != "" {
			req.TenantId = tenantID
		}
		if module != "" {
			req.Module = module
		}
		if locale != "" {
			req.Locale = locale
		}
		if len(codes) > 0 {
			req.Codes = codes
		}
	}

	// Validate the request
	if req.TenantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantId is required"})
		return
	}

	log.Printf("Searching messages with tenantId=%s, module=%s, locale=%s, codes=%v",
		req.TenantId, req.Module, req.Locale, req.Codes)

	var messages []domain.Message
	var err error

	// If codes are provided, search by codes
	if len(req.Codes) > 0 {
		messages, err = h.service.SearchMessagesByCodes(c.Request.Context(), req.TenantId, req.Locale, req.Codes)
	} else {
		// Otherwise search by module and locale
		messages, err = h.service.SearchMessages(c.Request.Context(), req.TenantId, req.Module, req.Locale)
	}

	if err != nil {
		log.Printf("Error searching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Found %d messages", len(messages))

	// Convert domain messages to response format
	responseMessages := make([]dtos.MessageResponse, len(messages))
	for i, msg := range messages {
		responseMessages[i] = dtos.MessageResponse{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	// Return the response
	c.JSON(http.StatusOK, dtos.SearchMessagesResponse{
		Messages: responseMessages,
	})
}

// RegisterRoutes registers the API routes on the given router group
func (h *MessageHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/localization/messages/v1/_upsert", h.UpsertMessages)
	router.POST("/localization/messages/v1/_search", h.SearchMessages)
}
