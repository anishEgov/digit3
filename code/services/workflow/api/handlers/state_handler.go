package handlers

import (
	"net/http"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

type StateHandler struct {
	service service.StateService
}

func NewStateHandler(s service.StateService) *StateHandler {
	return &StateHandler{service: s}
}

// CreateState handles the API request to create a new state for a process.
func (h *StateHandler) CreateState(c *gin.Context) {
	var state models.State
	if err := c.ShouldBindJSON(&state); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	state.ProcessID = c.Param("id") // Changed from "processId"
	state.TenantID = c.GetHeader("X-Tenant-ID")
	// state.AuditDetail.CreatedBy = ...

	createdState, err := h.service.CreateState(c.Request.Context(), &state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdState)
}

// GetStates handles the API request to list all states for a process.
func (h *StateHandler) GetStates(c *gin.Context) {
	processID := c.Param("id") // Changed from "processId"
	tenantID := c.GetHeader("X-Tenant-ID")

	states, err := h.service.GetStatesByProcessID(c.Request.Context(), tenantID, processID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, states)
}

// GetState handles the API request to get a single state by its ID.
func (h *StateHandler) GetState(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	state, err := h.service.GetStateByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{Code: "NotFound", Message: "State not found"})
		return
	}

	c.JSON(http.StatusOK, state)
}

// UpdateState handles the API request to update a state.
func (h *StateHandler) UpdateState(c *gin.Context) {
	id := c.Param("id")
	var state models.State
	if err := c.ShouldBindJSON(&state); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	state.ID = id
	state.TenantID = c.GetHeader("X-Tenant-ID")
	// state.AuditDetail.ModifiedBy = ...

	updatedState, err := h.service.UpdateState(c.Request.Context(), &state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedState)
}

// DeleteState handles the API request to delete a state.
func (h *StateHandler) DeleteState(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	err := h.service.DeleteState(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
