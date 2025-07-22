package models

// AuditDetail represents the audit fields for all models.
type AuditDetail struct {
	CreatedBy    string `json:"createdBy" db:"created_by"`
	CreatedTime  int64  `json:"createdTime" db:"created_at"`
	ModifiedBy   string `json:"modifiedBy,omitempty" db:"modified_by"`
	ModifiedTime int64  `json:"modifiedTime,omitempty" db:"modified_at"`
}

// Document represents a document object from the common specification.
type Document struct {
	ID           string `json:"id,omitempty" db:"id"`
	DocumentType string `json:"documentType" db:"document_type"`
	FileStoreID  string `json:"fileStoreId" db:"file_store_id"`
	DocumentUID  string `json:"documentUid,omitempty" db:"document_uid"`
}

// Error represents a single API error message.
type Error struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Description string   `json:"description,omitempty"`
	Params      []string `json:"params,omitempty"`
}
