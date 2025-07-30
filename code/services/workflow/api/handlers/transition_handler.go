package handlers

import (
	"context"
	"net/http"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

type TransitionHandler struct {
	transitionService service.TransitionService
}

func NewTransitionHandler(transitionService service.TransitionService) *TransitionHandler {
	return &TransitionHandler{
		transitionService: transitionService,
	}
}

type TransitionRequest struct {
	ProcessInstanceID *string             `json:"processInstanceId,omitempty"`
	ProcessID         string              `json:"processId" binding:"required"`
	EntityID          string              `json:"entityId" binding:"required"`
	Action            string              `json:"action" binding:"required"`
	Comment           *string             `json:"comment,omitempty"`
	Documents         []models.Document   `json:"documents,omitempty"`
	Assignees         *[]string           `json:"assignees,omitempty"`
	Attributes        map[string][]string `json:"attributes,omitempty"` // User attributes for validation
}

type TransitionResponse struct {
	ID           string              `json:"id"`
	ProcessID    string              `json:"processId"`
	EntityID     string              `json:"entityId"`
	Action       string              `json:"action"`
	Status       string              `json:"status"`
	Comment      *string             `json:"comment,omitempty"`
	Documents    []models.Document   `json:"documents,omitempty"`
	Assigner     *string             `json:"assigner,omitempty"`
	Assignees    []string            `json:"assignees,omitempty"`
	CurrentState string              `json:"currentState"`
	StateSla     *int64              `json:"stateSla,omitempty"`
	ProcessSla   *int64              `json:"processSla,omitempty"`
	Attributes   map[string][]string `json:"attributes,omitempty"`
	AuditDetails models.AuditDetail  `json:"auditDetails"`
}

func (h *TransitionHandler) Transition(c *gin.Context) {
	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Extract user ID from X-Client-Id header for audit tracking
	userID := models.GetUserIDFromContext(c)

	// Add user information to context for the service layer
	ctx := context.WithValue(c.Request.Context(), "userID", userID)
	ctx = context.WithValue(ctx, "tenantID", tenantID)

	// Call the transition service
	result, err := h.transitionService.Transition(ctx, req.ProcessInstanceID, req.ProcessID, req.EntityID, req.Action, req.Comment, req.Documents, req.Assignees, req.Attributes, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := TransitionResponse{
		ID:           result.ID,
		ProcessID:    result.ProcessID,
		EntityID:     result.EntityID,
		Action:       req.Action,
		Status:       result.Status,
		Comment:      req.Comment,
		Documents:    result.Documents,
		Assigner:     result.Assigner,
		Assignees:    result.Assignees,
		CurrentState: result.CurrentState,
		StateSla:     result.StateSLA,   // Correct field name
		ProcessSla:   result.ProcessSLA, // Correct field name
		Attributes:   result.Attributes,
		AuditDetails: result.AuditDetails, // Correct field name
	}

	c.JSON(http.StatusOK, response)
}

func (h *TransitionHandler) GetTransitions(c *gin.Context) {
	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Get query parameters
	entityID := c.Query("entityId")
	processID := c.Query("processId")
	historyParam := c.DefaultQuery("history", "false")

	// Validate required parameters
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entityId query parameter is required"})
		return
	}
	if processID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "processId query parameter is required"})
		return
	}

	// Parse history parameter
	history := historyParam == "true"

	// Add tenant information to context
	ctx := context.WithValue(c.Request.Context(), "tenantID", tenantID)

	// Call the service
	instances, err := h.transitionService.GetTransitions(ctx, tenantID, entityID, processID, history)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"processInstances": instances,
		"totalCount":       len(instances),
	})
}
