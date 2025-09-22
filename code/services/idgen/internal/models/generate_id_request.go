package models

type GenerateIDRequest struct {
	TemplateID string            `json:"templateId" binding:"required"`
	Variables  map[string]string `json:"variables"`
}
