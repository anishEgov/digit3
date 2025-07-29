package models

import (
	"time"

	"github.com/gin-gonic/gin"
)

// AuditDetail represents audit information for database records.
type AuditDetail struct {
	CreatedBy    string `json:"createdBy,omitempty" db:"created_by"`
	CreatedTime  int64  `json:"createdTime,omitempty" db:"created_at"`
	ModifiedBy   string `json:"modifiedBy,omitempty" db:"modified_by"`
	ModifiedTime int64  `json:"modifiedTime,omitempty" db:"modified_at"`
}

// GetUserIDFromContext extracts user ID from X-Client-Id header with fallback to "system"
func GetUserIDFromContext(c *gin.Context) string {
	clientID := c.GetHeader("X-Client-Id")
	if clientID == "" {
		return "system" // Fallback if no client ID provided
	}
	return clientID
}

// SetAuditDetailsForCreate sets audit details for a new record creation
func (a *AuditDetail) SetAuditDetailsForCreate(userID string) {
	now := time.Now().UnixMilli()
	a.CreatedBy = userID
	a.CreatedTime = now
	a.ModifiedBy = userID
	a.ModifiedTime = now
}

// SetAuditDetailsForUpdate sets audit details for record update
func (a *AuditDetail) SetAuditDetailsForUpdate(userID string) {
	now := time.Now().UnixMilli()
	a.ModifiedBy = userID
	a.ModifiedTime = now
}

// Document represents a document attachment.
type Document struct {
	DocumentType      string                 `json:"documentType,omitempty" db:"document_type"`
	FileStoreID       string                 `json:"fileStoreId,omitempty" db:"file_store_id"`
	DocumentUID       string                 `json:"documentUid,omitempty" db:"document_uid"`
	AdditionalDetails map[string]interface{} `json:"additionalDetails,omitempty" db:"additional_details"`
}

// Error represents an API error response.
type Error struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}
