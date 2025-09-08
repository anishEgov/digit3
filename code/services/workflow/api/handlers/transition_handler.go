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
	Init              *bool               `json:"init,omitempty"` // NEW: Flag to create new instance in initial state
	Action            string              `json:"action"`
	Status            *string             `json:"status,omitempty"`
	CurrentState      *string             `json:"currentState,omitempty"` // Expected current state for validation
	Comment           *string             `json:"comment,omitempty"`
	Documents         []string            `json:"documents,omitempty"`
	Assigner          *string             `json:"assigner,omitempty"`
	Assignees         *[]string           `json:"assignees,omitempty"`
	Attributes        map[string][]string `json:"attributes,omitempty"` // User attributes for validation
}

type TransitionResponse struct {
	ID           string              `json:"id"`
	ProcessID    string              `json:"processId"`
	EntityID     string              `json:"entityId"`
	Action       string              `json:"action"`
	Status       string              `json:"status"`
	Comment      string              `json:"comment"`
	Documents    []string            `json:"documents"`
	Assigner     string              `json:"assigner"`
	Assignees    []string            `json:"assignees"`
	CurrentState string              `json:"currentState"`
	StateSla     int64               `json:"stateSla"`
	ProcessSla   int64               `json:"processSla"`
	Attributes   map[string][]string `json:"attributes"`
	AuditDetails models.AuditDetail  `json:"auditDetails"`
}

func (h *TransitionHandler) Transition(c *gin.Context) {
	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// No validation needed - the old logic handles both creation and transition cases

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

	// Call the transition service with init parameter
	result, err := h.transitionService.Transition(ctx, req.ProcessInstanceID, req.ProcessID, req.EntityID, req.Action, req.Init, req.Status, req.CurrentState, req.Comment, req.Documents, req.Assigner, req.Assignees, req.Attributes, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert documents from []models.Document to []string
	var docStrings []string
	for _, doc := range result.Documents {
		docStrings = append(docStrings, doc.FileStoreID)
	}

	// Handle optional fields with defaults
	comment := ""
	if req.Comment != nil {
		comment = *req.Comment
	}

	assigner := ""
	if result.Assigner != nil {
		assigner = *result.Assigner
	}

	stateSla := int64(0)
	if result.StateSLA != nil {
		stateSla = *result.StateSLA
	}

	processSla := int64(0)
	if result.ProcessSLA != nil {
		processSla = *result.ProcessSLA
	}

	// Ensure attributes is not nil
	attributes := result.Attributes
	if attributes == nil {
		attributes = make(map[string][]string)
	}

	response := TransitionResponse{
		ID:           result.ID,
		ProcessID:    result.ProcessID,
		EntityID:     result.EntityID,
		Action:       req.Action,
		Status:       result.Status,
		Comment:      comment,
		Documents:    docStrings,
		Assigner:     assigner,
		Assignees:    result.Assignees,
		CurrentState: result.CurrentState,
		StateSla:     stateSla,
		ProcessSla:   processSla,
		Attributes:   attributes,
		AuditDetails: result.AuditDetails,
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
