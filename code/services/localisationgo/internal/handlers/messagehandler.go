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

// CreateMessages handles the messages create API endpoint
func (h *MessageHandler) CreateMessages(c *gin.Context) {
	var req dtos.CreateMessagesRequest
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
	messages, err := h.service.CreateMessages(c.Request.Context(), req.TenantId, req.Messages)
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
	c.JSON(http.StatusOK, dtos.CreateMessagesResponse{
		Messages: responseMessages,
	})
}

// UpdateMessages handles the messages update API endpoint
func (h *MessageHandler) UpdateMessages(c *gin.Context) {
	var req dtos.UpdateMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if req.TenantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantId is required"})
		return
	}
	if req.Locale == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "locale is required"})
		return
	}
	if req.Module == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module is required"})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one message is required"})
		return
	}

	// Check for duplicate codes within the request
	codeMap := make(map[string]bool)
	for i, msg := range req.Messages {
		if _, exists := codeMap[msg.Code]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("duplicate message code at index %d: %s", i, msg.Code)})
			return
		}
		codeMap[msg.Code] = true
	}

	// Convert update messages to domain messages
	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  req.Module,
			Locale:  req.Locale,
		}
	}

	// Call the service
	messages, err := h.service.UpdateMessagesForModule(c.Request.Context(), req.TenantId, req.Locale, req.Module, domainMessages)
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
	c.JSON(http.StatusOK, dtos.UpdateMessagesResponse{
		Messages: responseMessages,
	})
}

// DeleteMessages handles the messages delete API endpoint
func (h *MessageHandler) DeleteMessages(c *gin.Context) {
	var req dtos.DeleteMessagesRequest
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

	// Convert delete messages to message identities
	messageIdentities := make([]dtos.MessageIdentity, len(req.Messages))
	for i, msg := range req.Messages {
		messageIdentities[i] = dtos.MessageIdentity{
			TenantId: req.TenantId,
			Module:   msg.Module,
			Locale:   msg.Locale,
			Code:     msg.Code,
		}
	}

	// Call the service
	err := h.service.DeleteMessages(c.Request.Context(), messageIdentities)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, dtos.DeleteMessagesResponse{
		Success: true,
	})
}

// BustCache handles the cache bust API endpoint
func (h *MessageHandler) BustCache(c *gin.Context) {
	// Call the service
	err := h.service.BustCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, dtos.CacheBustResponse{
		Message: "Cache cleared successfully",
		Success: true,
	})
}

// SearchMessages handles the messages search API endpoint
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	var req dtos.SearchMessagesRequest

	// For GET requests, get parameters from URL query
	tenantID := c.Query("tenantId")
	module := c.Query("module")
	locale := c.Query("locale")
	codeStr := c.Query("codes")

	// Create RequestInfo with minimal data
	req.RequestInfo = models.RequestInfo{
		APIId: "api.localization",
		Ver:   "1.0",
		Ts:    time.Now(),
		MsgId: uuid.New().String(),
	}

	// Update request with query parameters
	req.TenantId = tenantID
	req.Module = module
	req.Locale = locale
	if codeStr != "" {
		req.Codes = strings.Split(codeStr, ",")
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
	// Search/Read operations - use GET
	router.GET("/localization/messages/v1/_search", h.SearchMessages)
	router.GET("/localization/messages", h.SearchMessages) // URL parameter-based search

	// Create operation - use POST
	router.POST("/localization/messages/v1/_create", h.CreateMessages)

	// Update operations - use PUT
	router.PUT("/localization/messages/v1/_upsert", h.UpsertMessages)
	router.PUT("/localization/messages/v1/_update", h.UpdateMessages)

	// Delete operations - use DELETE
	router.DELETE("/localization/messages/v1/_delete", h.DeleteMessages)
	router.DELETE("/localization/messages/cache-bust", h.BustCache)
}
