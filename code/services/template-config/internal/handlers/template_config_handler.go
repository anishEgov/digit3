package handlers

import (
	"net/http"
	"strings"
	"template-config/internal/models"
	"template-config/internal/service"
	"template-config/internal/validation"

	"github.com/gin-gonic/gin"
)

type TemplateConfigHandler struct {
	service   *service.TemplateConfigService
	validator *validation.TemplateValidator
}

func NewTemplateConfigHandler(service *service.TemplateConfigService) *TemplateConfigHandler {
	return &TemplateConfigHandler{service: service, validator: validation.NewTemplateValidator()}
}

func getTenantIDFromHeader(c *gin.Context) string {
	return c.GetHeader("X-Tenant-ID")
}

// CreateTemplateConfig handles POST /template-config
func (h *TemplateConfigHandler) CreateTemplateConfig(c *gin.Context) {
	var config models.TemplateConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	config.TenantID = getTenantIDFromHeader(c)

	// Validate full template config (fieldMapping, apiMapping, etc.)
	if err := h.validator.ValidateTemplateConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid template config",
			Description: err.Error(),
		})
		return
	}

	dbConfig := models.FromDTO(&config)
	if err := h.service.Create(&dbConfig); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, models.Error{
				Code:        "CONFLICT",
				Message:     "Template config already exists",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to create template config",
			Description: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, dbConfig.ToDTO())
}

// UpdateTemplateConfig handles PUT /template-config
func (h *TemplateConfigHandler) UpdateTemplateConfig(c *gin.Context) {
	var config models.TemplateConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	config.TenantID = getTenantIDFromHeader(c)

	// Validate full template config (fieldMapping, apiMapping, etc.)
	if err := h.validator.ValidateTemplateConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid template config",
			Description: err.Error(),
		})
		return
	}

	dbConfig := models.FromDTO(&config)
	if err := h.service.Update(&dbConfig); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusNotFound, models.Error{
				Code:        "NOT_FOUND",
				Message:     "Template config not found",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to update template config",
			Description: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dbConfig.ToDTO())
}

// SearchTemplateConfigs handles GET /template-config
func (h *TemplateConfigHandler) SearchTemplateConfigs(c *gin.Context) {
	var search models.TemplateConfigSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid query parameters",
			Description: err.Error(),
		})
		return
	}
	if uuidsStr := c.Query("uuids"); uuidsStr != "" {
		search.IDs = strings.Split(uuidsStr, ",")
	}
	search.TenantID = getTenantIDFromHeader(c)
	configs, err := h.service.Search(&search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to search template configs",
			Description: err.Error(),
		})
		return
	}
	// Map to API models
	var apiConfigs []models.TemplateConfig
	for _, config := range configs {
		apiConfigs = append(apiConfigs, config.ToDTO())
	}
	c.JSON(http.StatusOK, apiConfigs)
}

// DeleteTemplateConfig handles DELETE /template-config
func (h *TemplateConfigHandler) DeleteTemplateConfig(c *gin.Context) {
	var deleteReq models.TemplateConfigDelete
	if err := c.ShouldBindQuery(&deleteReq); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid query parameters",
			Description: err.Error(),
		})
		return
	}
	deleteReq.TenantID = getTenantIDFromHeader(c)
	if err := h.service.Delete(deleteReq.TemplateID, deleteReq.TenantID, deleteReq.Version); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusNotFound, models.Error{
				Code:        "NOT_FOUND",
				Message:     "Template config not found",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to delete template config",
			Description: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

// RenderTemplateConfig handles POST /template-config/render
func (h *TemplateConfigHandler) RenderTemplateConfig(c *gin.Context) {
	var request models.RenderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	request.TenantID = getTenantIDFromHeader(c)
	response, errors := h.service.Render(&request)
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, errors)
		return
	}
	c.JSON(http.StatusOK, response)
}
