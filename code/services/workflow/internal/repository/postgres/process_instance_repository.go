package postgres

import (
	"context"
	"encoding/json"
	"time"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type processInstanceRepository struct {
	db *sqlx.DB
}

func NewProcessInstanceRepository(db *sqlx.DB) repository.ProcessInstanceRepository {
	return &processInstanceRepository{db: db}
}

func generateUUID() string {
	return uuid.New().String()
}

func (r *processInstanceRepository) CreateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error {
	if instance.ID == "" {
		instance.ID = generateUUID()
	}

	// Set audit details
	now := time.Now().UnixMilli()
	instance.AuditDetails.CreatedTime = now
	instance.AuditDetails.ModifiedTime = now
	if instance.AuditDetails.CreatedBy == "" {
		instance.AuditDetails.CreatedBy = "system"
	}
	if instance.AuditDetails.ModifiedBy == "" {
		instance.AuditDetails.ModifiedBy = "system"
	}

	// Marshal JSON fields
	documentsJSON, _ := json.Marshal(instance.Documents)
	assigneesJSON, _ := json.Marshal(instance.Assignees)
	attributesJSON, _ := json.Marshal(instance.Attributes)

	query := `INSERT INTO process_instances (id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, created_by, created_at, modified_by, modified_at)
              VALUES (:id, :tenant_id, :process_id, :entity_id, :action, :status, :comment, :documents, :assigner, :assignees, :current_state_id, :state_sla, :process_sla, :attributes, :created_by, :created_at, :modified_by, :modified_at)`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":               instance.ID,
		"tenant_id":        instance.TenantID,
		"process_id":       instance.ProcessID,
		"entity_id":        instance.EntityID,
		"action":           instance.Action,
		"status":           instance.Status,
		"comment":          instance.Comment,
		"documents":        documentsJSON,
		"assigner":         instance.Assigner,
		"assignees":        assigneesJSON,
		"current_state_id": instance.CurrentState,
		"state_sla":        instance.StateSLA,
		"process_sla":      instance.ProcessSLA,
		"attributes":       attributesJSON,
		"created_by":       instance.AuditDetails.CreatedBy,
		"created_at":       instance.AuditDetails.CreatedTime,
		"modified_by":      instance.AuditDetails.ModifiedBy,
		"modified_at":      instance.AuditDetails.ModifiedTime,
	})
	return err
}

func (r *processInstanceRepository) GetProcessInstanceByID(ctx context.Context, tenantID, id string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	query := `SELECT * FROM process_instances WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &instance, query, tenantID, id)
	return &instance, err
}

func (r *processInstanceRepository) GetProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	query := `SELECT * FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3`
	err := r.db.GetContext(ctx, &instance, query, tenantID, entityID, processID)
	return &instance, err
}

func (r *processInstanceRepository) UpdateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error {
	// Marshal JSON fields
	documentsJSON, _ := json.Marshal(instance.Documents)
	assigneesJSON, _ := json.Marshal(instance.Assignees)
	attributesJSON, _ := json.Marshal(instance.Attributes)

	// Update audit details
	now := time.Now().UnixMilli()
	instance.AuditDetails.ModifiedTime = now
	if instance.AuditDetails.ModifiedBy == "" {
		instance.AuditDetails.ModifiedBy = "system"
	}

	query := `UPDATE process_instances 
			  SET action = :action,
				  status = :status,
				  comment = :comment,
				  documents = :documents,
				  assigner = :assigner,
				  assignees = :assignees,
				  current_state_id = :current_state_id,
				  state_sla = :state_sla,
				  process_sla = :process_sla,
				  attributes = :attributes,
				  modified_by = :modified_by,
				  modified_at = :modified_at
			  WHERE tenant_id = :tenant_id AND id = :id`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":               instance.ID,
		"tenant_id":        instance.TenantID,
		"action":           instance.Action,
		"status":           instance.Status,
		"comment":          instance.Comment,
		"documents":        documentsJSON,
		"assigner":         instance.Assigner,
		"assignees":        assigneesJSON,
		"current_state_id": instance.CurrentState,
		"state_sla":        instance.StateSLA,
		"process_sla":      instance.ProcessSLA,
		"attributes":       attributesJSON,
		"modified_by":      instance.AuditDetails.ModifiedBy,
		"modified_at":      instance.AuditDetails.ModifiedTime,
	})
	return err
}
