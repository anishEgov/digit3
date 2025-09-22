package handlers

import (
	"idgen/internal/models"
	"idgen/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type IDGenHandler struct {
	svc *service.IDGenService
}

func NewIDGenHandler(svc *service.IDGenService) *IDGenHandler {
	return &IDGenHandler{svc: svc}
}

func getClientIDFromHeader(c *gin.Context) string {
	return c.GetHeader("X-Client-ID")
}

func (h *IDGenHandler) RegisterTemplate(c *gin.Context) {
	var req models.RegisterTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Params:  []string{err.Error()},
		})
		return
	}

	req.CreatedBy = getClientIDFromHeader(c)
	req.CreatedTime = time.Now().Unix()

	if err := h.svc.RegisterTemplate(req); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to register template",
			Params:  []string{err.Error()},
		})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *IDGenHandler) GenerateID(c *gin.Context) {
	var req models.GenerateIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Params:  []string{err.Error()},
		})
		return
	}

	id, err := h.svc.GenerateID(req.TemplateID, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    "GENERATION_FAILED",
			Message: "Failed to generate ID",
			Params:  []string{err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.GenerateIDResponse{ID: id})
}
