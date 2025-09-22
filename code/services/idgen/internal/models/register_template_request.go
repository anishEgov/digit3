package models

type RegisterTemplateRequest struct {
	TemplateID  string              `json:"templateId" binding:"required"`
	Config      IDGenTemplateConfig `json:"config" binding:"required"`
	CreatedBy   string              `json:"createdBy"`
	CreatedTime int64               `json:"createdTime"`
}

type IDGenTemplateConfig struct {
	Template string         `json:"template" binding:"required"`
	Sequence SequenceConfig `json:"sequence"`
	Random   RandomConfig   `json:"random"`
}

type SequenceConfig struct {
	Scope   string        `json:"scope" default:"global"`
	Start   int           `json:"start" default:"1"`
	Padding PaddingConfig `json:"padding"`
}

type PaddingConfig struct {
	Length int    `json:"length"`
	Char   string `json:"char" default:"0"`
}

type RandomConfig struct {
	Length  int    `json:"length" default:"0"`
	Charset string `json:"charset"`
}
