package models

// AuditDetails represents the audit information for entities
// Use string for createdBy/lastModifiedBy and int64 for createdTime/lastModifiedTime
type AuditDetails struct {
	CreatedBy        string `json:"createdBy,omitempty" gorm:"column:createdby"`
	LastModifiedBy   string `json:"lastModifiedBy,omitempty" gorm:"column:lastmodifiedby"`
	CreatedTime      int64  `json:"createdTime,omitempty" gorm:"column:createdtime"`
	LastModifiedTime int64  `json:"lastModifiedTime,omitempty" gorm:"column:lastmodifiedtime"`
}

// ResponseInfo represents the response information for API responses
type ResponseInfo struct {
	APIId    string `json:"apiId,omitempty"`
	Ver      string `json:"ver,omitempty"`
	Ts       int64  `json:"ts,omitempty"`
	ResMsgId string `json:"resMsgId,omitempty"`
	MsgId    string `json:"msgId,omitempty"`
	Status   string `json:"status,omitempty"`
}

type Error struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Description string   `json:"description"`
	Params      []string `json:"params"`
}

type ErrorResponse struct {
	ResponseInfo ResponseInfo `json:"ResponseInfo"`
	Errors       []Error      `json:"Errors"`
} 