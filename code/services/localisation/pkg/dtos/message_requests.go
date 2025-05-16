package dtos

import (
	"localisationgo/internal/common/models"
	"localisationgo/internal/core/domain"
)

// UpsertMessagesRequest represents the request for upserting localization messages
type UpsertMessagesRequest struct {
	RequestInfo models.RequestInfo `json:"RequestInfo"`
	TenantId    string             `json:"tenantId"`
	Messages    []domain.Message   `json:"messages"`
}

// SearchMessagesRequest represents the request for searching localization messages
type SearchMessagesRequest struct {
	RequestInfo models.RequestInfo `json:"RequestInfo"`
	TenantId    string             `json:"tenantId"`
	Module      string             `json:"module,omitempty"`
	Locale      string             `json:"locale,omitempty"`
	Codes       []string           `json:"codes,omitempty"`
}
