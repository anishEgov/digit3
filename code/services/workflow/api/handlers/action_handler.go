package handlers

import (
	"net/http"
	"strings"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

// ActionHandler handles action-related HTTP requests.
type ActionHandler struct {
	actionService service.ActionService
}

// NewActionHandler creates a new ActionHandler.
func NewActionHandler(actionService service.ActionService) *ActionHandler {
	return &ActionHandler{
		actionService: actionService,
	}
}

// CreateAction handles the creation of a new action.
func (h *ActionHandler) CreateAction(c *gin.Context) {
	var action models.Action

	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	action.TenantID = tenantID

	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdAction, err := h.actionService.CreateAction(c.Request.Context(), &action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdAction)
}

// GetActions handles retrieving actions by state ID.
func (h *ActionHandler) GetActions(c *gin.Context) {
	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	stateID := c.Param("id")
	actions, err := h.actionService.GetActionsByStateID(c.Request.Context(), tenantID, stateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, actions)
}

// GetAction handles retrieving a single action by ID.
func (h *ActionHandler) GetAction(c *gin.Context) {
	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	actionID := c.Param("id")
	action, err := h.actionService.GetActionByID(c.Request.Context(), tenantID, actionID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Action not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, action)
}

// UpdateAction handles updating an existing action.
func (h *ActionHandler) UpdateAction(c *gin.Context) {
	var action models.Action

	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	action.TenantID = tenantID

	actionID := c.Param("id")
	action.ID = actionID

	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAction, err := h.actionService.UpdateAction(c.Request.Context(), &action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedAction)
}

// DeleteAction handles deleting an action.
func (h *ActionHandler) DeleteAction(c *gin.Context) {
	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	actionID := c.Param("id")
	err := h.actionService.DeleteAction(c.Request.Context(), tenantID, actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
