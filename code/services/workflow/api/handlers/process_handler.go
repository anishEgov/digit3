package handlers

import (
	"net/http"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/service"
	"github.com/gin-gonic/gin"
)

type ProcessHandler struct {
	service service.ProcessService
}

func NewProcessHandler(s service.ProcessService) *ProcessHandler {
	return &ProcessHandler{service: s}
}

// CreateProcess handles the API request to create a new process.
func (h *ProcessHandler) CreateProcess(c *gin.Context) {
	var process models.Process
	if err := c.ShouldBindJSON(&process); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	// Extract tenantID and user info from headers/context
	process.TenantID = c.GetHeader("X-Tenant-ID")
	// process.AuditDetail.CreatedBy = ... (get from JWT or context)

	createdProcess, err := h.service.CreateProcess(c.Request.Context(), &process)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdProcess)
}

// GetProcess handles the API request to get a process by its ID.
func (h *ProcessHandler) GetProcess(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	process, err := h.service.GetProcessByID(c.Request.Context(), tenantID, id)
	if err != nil {
		// Differentiate between not found and other errors
		c.JSON(http.StatusNotFound, models.Error{Code: "NotFound", Message: "Process not found"})
		return
	}

	c.JSON(http.StatusOK, process)
}

// GetProcesses handles the API request to list/search for processes.
func (h *ProcessHandler) GetProcesses(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	ids := c.QueryArray("id")
	names := c.QueryArray("name")

	processes, err := h.service.GetProcesses(c.Request.Context(), tenantID, ids, names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, processes)
}

// GetProcessDefinitions handles the API request to list full process definitions.
func (h *ProcessHandler) GetProcessDefinitions(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	ids := c.QueryArray("id")
	names := c.QueryArray("name")

	definitions, err := h.service.GetProcessDefinitions(c.Request.Context(), tenantID, ids, names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, definitions)
}

// UpdateProcess handles the API request to update a process.
func (h *ProcessHandler) UpdateProcess(c *gin.Context) {
	id := c.Param("id")
	var process models.Process
	if err := c.ShouldBindJSON(&process); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	process.ID = id
	process.TenantID = c.GetHeader("X-Tenant-ID")
	// process.AuditDetail.ModifiedBy = ... (get from JWT or context)

	updatedProcess, err := h.service.UpdateProcess(c.Request.Context(), &process)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedProcess)
}

// DeleteProcess handles the API request to delete a process.
func (h *ProcessHandler) DeleteProcess(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")

	err := h.service.DeleteProcess(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
