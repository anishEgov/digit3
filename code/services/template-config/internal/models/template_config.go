package models

import (
	"github.com/google/uuid"
)

// TemplateConfig is the API request/response model with nested auditDetails
type TemplateConfig struct {
	ID           uuid.UUID         `json:"id"`
	TemplateID   string            `json:"templateId" binding:"required"`
	TenantID     string            `json:"tenantId"`
	Version      string            `json:"version" binding:"required"`
	FieldMapping map[string]string `json:"fieldMapping"`
	APIMapping   []APIMapping      `json:"apiMapping"`
	AuditDetails AuditDetails      `json:"auditDetails"`
}
