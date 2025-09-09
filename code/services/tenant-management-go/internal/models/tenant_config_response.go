package models

type TenantConfigRequest struct {
	TenantConfig TenantConfig `json:"tenantConfig"`
}

type TenantConfigResponse struct {
	ResponseInfo   ResponseInfo   `json:"ResponseInfo"`
	TenantConfigs  []TenantConfig `json:"tenantConfigs"`
} 