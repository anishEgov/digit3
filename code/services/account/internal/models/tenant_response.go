package models

type TenantRequest struct {
	Tenant Tenant `json:"tenant"`
}

type TenantResponse struct {
	ResponseInfo ResponseInfo `json:"ResponseInfo"`
	Tenants      []Tenant     `json:"tenants"`
} 