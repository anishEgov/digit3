package domain

import (
	"time"
)

// Message represents a localization message entry
type Message struct {
	ID               int64     `json:"-"` // Internal DB ID, hidden from public response
	UUID             string    `json:"uuid"`
	TenantID         string    `json:"tenantId"`
	Module           string    `json:"module"`
	Locale           string    `json:"locale"`
	Code             string    `json:"code"`
	Message          string    `json:"message"`
	CreatedBy        string    `json:"created_by,omitempty"`
	CreatedDate      time.Time `json:"created_date,omitempty"`
	LastModifiedBy   string    `json:"last_modified_by,omitempty"`
	LastModifiedDate time.Time `json:"last_modified_date,omitempty"`
}
