package enrichment

import (
	"time"
	"tenant-management-go/internal/models"
	"tenant-management-go/internal/util"
	"github.com/google/uuid"
)

func EnrichTenant(tenant *models.Tenant, clientID string) {
	now := time.Now().Unix()
	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}
	if tenant.Code == "" && tenant.Name != "" {
		tenant.Code = util.GenerateCodeFromName(tenant.Name)
	}
	tenant.CreatedBy = clientID
	tenant.LastModifiedBy = clientID
	tenant.CreatedTime = now
	tenant.LastModifiedTime = now
}

func EnrichTenantUpdate(tenant *models.Tenant, clientID string) {
	tenant.LastModifiedBy = clientID
	tenant.LastModifiedTime = time.Now().Unix()
} 