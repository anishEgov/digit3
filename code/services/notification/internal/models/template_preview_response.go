package models

// TemplatePreviewResponse represents the API response model for template preview
type TemplatePreviewResponse struct {
	Type            TemplateType `json:"type" binding:"required,oneof=EMAIL SMS"`
	IsHTML          bool         `json:"isHTML"`
	RenderedSubject string       `json:"renderedSubject,omitempty"`
	RenderedContent string       `json:"renderedContent"`
}
