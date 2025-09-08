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

	// Extract tenant ID from header (required for multi-tenancy)
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Validate input
	if validationErrors := models.ValidateProcessCreate(&process); validationErrors != nil {
		c.JSON(http.StatusBadRequest, validationErrors)
		return
	}

	// Extract user ID from X-Client-Id header and set audit details
	userID := models.GetUserIDFromContext(c)
	process.TenantID = tenantID
	process.AuditDetail.SetAuditDetailsForCreate(userID)

	createdProcess, err := h.service.CreateProcess(c.Request.Context(), &process)
	if err != nil {
		// Check for database constraint violations and return proper error codes
		if models.IsDatabaseConstraintError(err) {
			c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: models.GetConstraintErrorMessage(err)})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdProcess)
}

func (h *ProcessHandler) GetProcesses(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

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
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	ids := c.QueryArray("id")
	names := c.QueryArray("name")

	definitions, err := h.service.GetProcessDefinitions(c.Request.Context(), tenantID, ids, names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, definitions)
}

// GetProcess handles the API request to get a single process by ID.
func (h *ProcessHandler) GetProcess(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Validate UUID format
	if err := models.ValidateUUID(id, "processId"); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: err.Error()})
		return
	}

	process, err := h.service.GetProcessByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{Code: "NotFound", Message: "Process not found"})
		return
	}

	c.JSON(http.StatusOK, process)
}

// UpdateProcess handles the API request to update a process.
func (h *ProcessHandler) UpdateProcess(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Validate UUID format
	if err := models.ValidateUUID(id, "processId"); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: err.Error()})
		return
	}

	var updateRequest models.Process
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "BadRequest", Message: err.Error()})
		return
	}

	// Validate that at least one field is provided for update
	if updateRequest.Name == "" && updateRequest.Code == "" &&
		(updateRequest.Description == nil || *updateRequest.Description == "") &&
		(updateRequest.Version == nil || *updateRequest.Version == "") &&
		updateRequest.SLA == nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: "At least one field must be provided for update"})
		return
	}

	// First, get the existing process
	existingProcess, err := h.service.GetProcessByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{Code: "NotFound", Message: "Process not found"})
		return
	}

	// Extract user ID from X-Client-Id header and set audit details for update
	userID := models.GetUserIDFromContext(c)

	// Merge update request with existing process, only updating provided fields
	processToUpdate := *existingProcess
	if updateRequest.Name != "" {
		processToUpdate.Name = updateRequest.Name
	}
	if updateRequest.Code != "" {
		processToUpdate.Code = updateRequest.Code
	}
	if updateRequest.Description != nil && *updateRequest.Description != "" {
		processToUpdate.Description = updateRequest.Description
	}
	if updateRequest.Version != nil && *updateRequest.Version != "" {
		processToUpdate.Version = updateRequest.Version
	}
	if updateRequest.SLA != nil {
		processToUpdate.SLA = updateRequest.SLA
	}

	// Set audit details for update
	processToUpdate.AuditDetail.SetAuditDetailsForUpdate(userID)

	updatedProcess, err := h.service.UpdateProcess(c.Request.Context(), &processToUpdate)
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
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	// Validate UUID format
	if err := models.ValidateUUID(id, "processId"); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: err.Error()})
		return
	}

	err := h.service.DeleteProcess(c.Request.Context(), tenantID, id)
	if err != nil {
		// Check for database constraint violations and return proper error codes
		if models.IsDatabaseConstraintError(err) {
			c.JSON(http.StatusBadRequest, models.Error{Code: "ValidationError", Message: models.GetConstraintErrorMessage(err)})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{Code: "InternalServerError", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
