package handlers

import (
	"net/http"
	"notification/internal/models"
	"notification/internal/service"
	"notification/internal/validation"
	"strings"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	service   *service.TemplateService
	validator *validation.TemplateValidator
}

func NewTemplateHandler(service *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service, validator: validation.NewTemplateValidator()}
}

func getTenantIDFromHeader(c *gin.Context) string {
	return c.GetHeader("X-Tenant-ID")
}

// Create template handles POST /template
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var template models.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	template.TenantID = getTenantIDFromHeader(c)

	// Validate full template
	if err := h.validator.ValidateTemplate(&template); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid template",
			Description: err.Error(),
		})
		return
	}

	templateDB := models.FromDTO(&template)
	if err := h.service.Create(&templateDB); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, models.Error{
				Code:        "CONFLICT",
				Message:     "Template already exists",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to create template",
			Description: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, templateDB.ToDTO())
}

// UpdateTemplate handles PUT /template
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	var template models.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	template.TenantID = getTenantIDFromHeader(c)

	// Validate template
	if err := h.validator.ValidateTemplate(&template); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid template",
			Description: err.Error(),
		})
		return
	}

	templateDB := models.FromDTO(&template)
	if err := h.service.Update(&templateDB); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusNotFound, models.Error{
				Code:        "NOT_FOUND",
				Message:     "Template not found",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to update template",
			Description: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, templateDB.ToDTO())
}

// SearchTemplates handles GET /template
func (h *TemplateHandler) SearchTemplates(c *gin.Context) {
	var searchReq models.TemplateSearch
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid query parameters",
			Description: err.Error(),
		})
		return
	}

	searchReq.TenantID = getTenantIDFromHeader(c)
	templateDBList, err := h.service.Search(&searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to search templates",
			Description: err.Error(),
		})
		return
	}
	// Map to API models
	var templateList []models.Template
	for _, templateDB := range templateDBList {
		templateList = append(templateList, templateDB.ToDTO())
	}
	c.JSON(http.StatusOK, templateList)
}

// DeleteTemplate handles DELETE /template
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	var deleteReq models.TemplateDelete
	if err := c.ShouldBindQuery(&deleteReq); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid query parameters",
			Description: err.Error(),
		})
		return
	}
	deleteReq.TenantID = getTenantIDFromHeader(c)
	if err := h.service.Delete(&deleteReq); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusNotFound, models.Error{
				Code:        "NOT_FOUND",
				Message:     "Template not found",
				Description: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:        "INTERNAL_SERVER_ERROR",
			Message:     "Failed to delete template",
			Description: err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

// PreviewTemplate handles POST /template/preview
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	var request models.TemplatePreviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:        "BAD_REQUEST",
			Message:     "Invalid request body",
			Description: err.Error(),
		})
		return
	}
	request.TenantID = getTenantIDFromHeader(c)
	response, errors := h.service.Preview(&request)
	if len(errors) > 0 {
		c.JSON(http.StatusInternalServerError, errors)
		return
	}
	c.JSON(http.StatusOK, response)
}
