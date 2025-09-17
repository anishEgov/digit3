package internal

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterTemplateRequest is the request body for /template
type RegisterTemplateRequest struct {
	TemplateID string          `json:"templateId" binding:"required"`
	Config     json.RawMessage `json:"config" binding:"required"`
}

// GenerateIdRequest is the request body for generating an ID
type GenerateIdRequest struct {
	TemplateID string            `json:"templateId" binding:"required"`
	Variables  map[string]string `json:"variables"`
}

// GenerateIdResponse is the response for /generate
type GenerateIdResponse struct {
	ID string `json:"id"`
}

// RegisterTemplateHandler handles POST /template
func RegisterTemplateHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Store the template config in the DB
		_, err := db.Exec(
			`INSERT INTO idgen_templates (id, config, created_at) VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config`,
			req.TemplateID, req.Config, time.Now().Unix(),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store template: " + err.Error()})
			return
		}

		// TODO: Create sequence if needed (handled in full logic)

		c.JSON(http.StatusOK, gin.H{"message": "template registered"})
	}
}

// GenerateIdHandler handles POST /generate
func GenerateIdHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GenerateIdRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Fetch template config from DB
		var configRaw json.RawMessage
		err := db.QueryRow(`SELECT config FROM idgen_templates WHERE id = $1`, req.TemplateID).Scan(&configRaw)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template: " + err.Error()})
			return
		}

		var config TemplateConfig
		if err := json.Unmarshal(configRaw, &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid template config: " + err.Error()})
			return
		}

		id, err := GenerateIDFromConfig(db, req.TemplateID, &config, req.Variables)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Store the generated ID for audit/traceability
		variablesJSON, _ := json.Marshal(req.Variables)
		_, err = db.Exec(
			`INSERT INTO idgen_generated (template_id, generated_id, variables, created_at) VALUES ($1, $2, $3, $4)`,
			req.TemplateID, id, variablesJSON, time.Now().Unix(),
		)
		if err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to store generated ID: %v", err)
		}

		c.JSON(http.StatusOK, GenerateIdResponse{ID: id})
	}
} 