package models

import (
	"github.com/google/uuid"
)

// SMSCategory represents the classification of SMS notifications
type TemplateType string

const (
	TemplateTypeEmail TemplateType = "EMAIL"
	TemplateTypeSMS   TemplateType = "SMS"
)

// Template is the API request/response model with nested auditDetails
type Template struct {
	ID           uuid.UUID    `json:"id"`
	TemplateID   string       `json:"templateId" binding:"required"`
	TenantID     string       `json:"tenantId"`
	Version      string       `json:"version" binding:"required"`
	Type         TemplateType `json:"type" binding:"required,oneof=EMAIL SMS"`
	Subject      string       `json:"subject"`
	Content      string       `json:"content" binding:"required"`
	IsHTML       bool         `json:"isHTML"`
	AuditDetails AuditDetails `json:"auditDetails"`
}
