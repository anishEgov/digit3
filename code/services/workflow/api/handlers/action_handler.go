package handlers

import (
	"net/http"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

type ActionHandler struct {
	service service.ActionService
}

func NewActionHandler(s service.ActionService) *ActionHandler {
	return &ActionHandler{service: s}
}

// CreateAction handles the API request to create a new action for a state.
func (h *ActionHandler) CreateAction(c *gin.Context) {
	var action models.Action
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	action.CurrentState = c.Param("id") // Changed from "stateid"
	action.TenantID = c.GetHeader("X-Tenant-ID")
	// action.AuditDetail.CreatedBy = ...

	createdAction, err := h.service.CreateAction(c.Request.Context(), &action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdAction)
}

// GetActions handles the API request to list all actions for a state.
func (h *ActionHandler) GetActions(c *gin.Context) {
	stateID := c.Param("id") // Changed from "stateid"
	tenantID := c.GetHeader("X-Tenant-ID")

	actions, err := h.service.GetActionsByStateID(c.Request.Context(), tenantID, stateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, actions)
}

// GetAction handles the API request to get a single action by its ID.
func (h *ActionHandler) GetAction(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	action, err := h.service.GetActionByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{Code: "NotFound", Message: "Action not found"})
		return
	}

	c.JSON(http.StatusOK, action)
}

// UpdateAction handles the API request to update an action.
func (h *ActionHandler) UpdateAction(c *gin.Context) {
	id := c.Param("id")
	var action models.Action
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	action.ID = id
	action.TenantID = c.GetHeader("X-Tenant-ID")
	// action.AuditDetail.ModifiedBy = ...

	updatedAction, err := h.service.UpdateAction(c.Request.Context(), &action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedAction)
}

// DeleteAction handles the API request to delete an action.
func (h *ActionHandler) DeleteAction(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	err := h.service.DeleteAction(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
