package handlers

import (
	"context"
	"net/http"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/security"
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
	ProcessInstanceID *string           `json:"processInstanceId,omitempty"`
	EntityID          string            `json:"entityId" binding:"required"`
	ProcessCode       string            `json:"processCode" binding:"required"`
	Action            string            `json:"action" binding:"required"`
	Comment           *string           `json:"comment,omitempty"`
	Documents         []models.Document `json:"documents,omitempty"`
	Assignees         *[]string         `json:"assignees,omitempty"`
}

type TransitionResponse struct {
	ProcessInstanceID string      `json:"processInstanceId"`
	State             string      `json:"state"`
	Action            string      `json:"action"`
	NextActions       []string    `json:"nextActions"`
	AuditDetails      interface{} `json:"auditDetails"`
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

	// Extract JWT token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract user information from JWT token
	userInfo, err := security.ExtractUserInfoFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT token: " + err.Error()})
		return
	}

	// Add user information to context for the service layer
	ctx := context.WithValue(c.Request.Context(), "userID", userInfo.UserID)
	ctx = context.WithValue(ctx, "userRoles", userInfo.Roles)
	ctx = context.WithValue(ctx, "tenantID", tenantID)

	// Call the transition service
	result, err := h.transitionService.Transition(ctx, req.ProcessInstanceID, req.EntityID, req.ProcessCode, req.Action, req.Comment, req.Documents, req.Assignees, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := TransitionResponse{
		ProcessInstanceID: result.ID,
		State:             result.CurrentState,
		Action:            req.Action,
		NextActions:       result.NextActions,
		AuditDetails:      result.AuditDetails,
	}

	c.JSON(http.StatusOK, response)
}
