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

// CreateMessagesRequest represents the request for creating new localization messages
type CreateMessagesRequest struct {
	RequestInfo models.RequestInfo `json:"RequestInfo"`
	TenantId    string             `json:"tenantId"`
	Messages    []domain.Message   `json:"messages"`
}

// UpdateMessagesRequest represents the request for updating existing localization messages
type UpdateMessagesRequest struct {
	RequestInfo models.RequestInfo `json:"RequestInfo"`
	TenantId    string             `json:"tenantId"`
	Locale      string             `json:"locale"`
	Module      string             `json:"module"`
	Messages    []UpdateMessage    `json:"messages"`
}

// UpdateMessage represents a message to be updated
type UpdateMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DeleteMessage represents a message to be deleted
type DeleteMessage struct {
	Code   string `json:"code"`
	Module string `json:"module"`
	Locale string `json:"locale"`
}

// DeleteMessagesRequest represents the request for deleting localization messages
type DeleteMessagesRequest struct {
	RequestInfo models.RequestInfo `json:"RequestInfo"`
	TenantId    string             `json:"tenantId"`
	Messages    []DeleteMessage    `json:"messages"`
}

// MessageIdentity represents the unique identity of a message
type MessageIdentity struct {
	TenantId string `json:"tenantId"`
	Module   string `json:"module"`
	Locale   string `json:"locale"`
	Code     string `json:"code"`
}

// FindMissingMessagesRequest represents the request to find missing messages
type FindMissingMessagesRequest struct {
	Locales []string `json:"locales"`
}
