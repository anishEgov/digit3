package domain

import (
	"time"
)

// Message represents a localization message entry
type Message struct {
	ID               int64     `json:"id,omitempty"`
	TenantID         string    `json:"tenantId"`
	Module           string    `json:"module"`
	Locale           string    `json:"locale"`
	Code             string    `json:"code"`
	Message          string    `json:"message"`
	CreatedBy        int64     `json:"createdBy,omitempty"`
	CreatedDate      time.Time `json:"createdDate,omitempty"`
	LastModifiedBy   int64     `json:"lastModifiedBy,omitempty"`
	LastModifiedDate time.Time `json:"lastModifiedDate,omitempty"`
}
