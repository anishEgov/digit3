package postgres

import (
	"context"
	"encoding/json"
	"fmt"
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
		instance.ID = uuid.New().String()
	}

	// Marshal JSON fields for PostgreSQL
	documentsJSON, err := json.Marshal(instance.Documents)
	if err != nil {
		return fmt.Errorf("error marshaling documents: %w", err)
	}

	assigneesJSON, err := json.Marshal(instance.Assignees)
	if err != nil {
		return fmt.Errorf("error marshaling assignees: %w", err)
	}

	attributesJSON, err := json.Marshal(instance.Attributes)
	if err != nil {
		return fmt.Errorf("error marshaling attributes: %w", err)
	}

	// SQL query with parallel workflow fields and escalated field
	query := `INSERT INTO process_instances (id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, escalated, created_by, created_at, modified_by, modified_at)
	VALUES (:id, :tenant_id, :process_id, :entity_id, :action, :status, :comment, :documents, :assigner, :assignees, :current_state_id, :state_sla, :process_sla, :attributes, :parent_instance_id, :branch_id, :is_parallel_branch, :escalated, :created_by, :created_at, :modified_by, :modified_at)`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                 instance.ID,
		"tenant_id":          instance.TenantID,
		"process_id":         instance.ProcessID,
		"entity_id":          instance.EntityID,
		"action":             instance.Action,
		"status":             instance.Status,
		"comment":            instance.Comment,
		"documents":          documentsJSON,
		"assigner":           instance.Assigner,
		"assignees":          assigneesJSON,
		"current_state_id":   instance.CurrentState,
		"state_sla":          instance.StateSLA,
		"process_sla":        instance.ProcessSLA,
		"attributes":         attributesJSON,
		"parent_instance_id": instance.ParentInstanceID,
		"branch_id":          instance.BranchID,
		"is_parallel_branch": instance.IsParallelBranch,
		"escalated":          instance.Escalated,
		"created_by":         instance.AuditDetails.CreatedBy,
		"created_at":         instance.AuditDetails.CreatedTime,
		"modified_by":        instance.AuditDetails.ModifiedBy,
		"modified_at":        instance.AuditDetails.ModifiedTime,
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

func (r *processInstanceRepository) GetLatestProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	query := `SELECT id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, created_by, created_at, modified_by, modified_at 
		FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND is_parallel_branch = false ORDER BY created_at DESC LIMIT 1`
	fmt.Printf("🔍 REPO DEBUG: Getting latest instance for entityID=%s, processID=%s\n", entityID, processID)

	row := r.db.QueryRowContext(ctx, query, tenantID, entityID, processID)

	var documentsJSON, assigneesJSON, attributesJSON []byte
	err := row.Scan(
		&instance.ID,
		&instance.TenantID,
		&instance.ProcessID,
		&instance.EntityID,
		&instance.Action,
		&instance.Status,
		&instance.Comment,
		&documentsJSON,
		&instance.Assigner,
		&assigneesJSON,
		&instance.CurrentState,
		&instance.StateSLA,
		&instance.ProcessSLA,
		&attributesJSON,
		&instance.ParentInstanceID,
		&instance.BranchID,
		&instance.IsParallelBranch,
		&instance.AuditDetails.CreatedBy,
		&instance.AuditDetails.CreatedTime,
		&instance.AuditDetails.ModifiedBy,
		&instance.AuditDetails.ModifiedTime,
	)

	if err == nil {
		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		fmt.Printf("✅ REPO DEBUG: Found latest instance ID=%s, currentState=%s, action=%s, createdAt=%d\n",
			instance.ID, instance.CurrentState, instance.Action, instance.AuditDetails.CreatedTime)
	} else {
		fmt.Printf("❌ REPO DEBUG: Error getting latest instance: %v\n", err)
	}
	return &instance, err
}

func (r *processInstanceRepository) GetProcessInstancesByEntityID(ctx context.Context, tenantID, entityID, processID string, history bool) ([]*models.ProcessInstance, error) {
	var query string

	if history {
		// Return all records ordered by created_at (oldest first for chronological order)
		query = `SELECT id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, created_by, created_at, modified_by, modified_at 
			FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 ORDER BY created_at ASC`
	} else {
		// Return only the latest record
		query = `SELECT id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, created_by, created_at, modified_by, modified_at 
			FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 ORDER BY created_at DESC LIMIT 1`
	}

	rows, err := r.db.QueryContext(ctx, query, tenantID, entityID, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*models.ProcessInstance
	for rows.Next() {
		var instance models.ProcessInstance
		var documentsJSON, assigneesJSON, attributesJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessID,
			&instance.EntityID,
			&instance.Action,
			&instance.Status,
			&instance.Comment,
			&documentsJSON,
			&instance.Assigner,
			&assigneesJSON,
			&instance.CurrentState,
			&instance.StateSLA,
			&instance.ProcessSLA,
			&attributesJSON,
			&instance.ParentInstanceID,
			&instance.BranchID,
			&instance.IsParallelBranch,
			&instance.AuditDetails.CreatedBy,
			&instance.AuditDetails.CreatedTime,
			&instance.AuditDetails.ModifiedBy,
			&instance.AuditDetails.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
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

// GetActiveParallelInstances returns all active parallel branch instances for an entity
func (r *processInstanceRepository) GetActiveParallelInstances(ctx context.Context, tenantID, entityID, processID string) ([]*models.ProcessInstance, error) {
	query := `SELECT id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, created_by, created_at, modified_by, modified_at 
		FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND is_parallel_branch = true AND status != 'COMPLETED' ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, entityID, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*models.ProcessInstance
	for rows.Next() {
		var instance models.ProcessInstance
		var documentsJSON, assigneesJSON, attributesJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessID,
			&instance.EntityID,
			&instance.Action,
			&instance.Status,
			&instance.Comment,
			&documentsJSON,
			&instance.Assigner,
			&assigneesJSON,
			&instance.CurrentState,
			&instance.StateSLA,
			&instance.ProcessSLA,
			&attributesJSON,
			&instance.ParentInstanceID,
			&instance.BranchID,
			&instance.IsParallelBranch,
			&instance.AuditDetails.CreatedBy,
			&instance.AuditDetails.CreatedTime,
			&instance.AuditDetails.ModifiedBy,
			&instance.AuditDetails.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

// GetInstancesByBranch returns instances for a specific parallel branch
func (r *processInstanceRepository) GetInstancesByBranch(ctx context.Context, tenantID, entityID, processID, branchID string) ([]*models.ProcessInstance, error) {
	query := `SELECT id, tenant_id, process_id, entity_id, action, status, comment, documents, assigner, assignees, current_state_id, state_sla, process_sla, attributes, parent_instance_id, branch_id, is_parallel_branch, created_by, created_at, modified_by, modified_at 
		FROM process_instances WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND branch_id = $4 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, entityID, processID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*models.ProcessInstance
	for rows.Next() {
		var instance models.ProcessInstance
		var documentsJSON, assigneesJSON, attributesJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessID,
			&instance.EntityID,
			&instance.Action,
			&instance.Status,
			&instance.Comment,
			&documentsJSON,
			&instance.Assigner,
			&assigneesJSON,
			&instance.CurrentState,
			&instance.StateSLA,
			&instance.ProcessSLA,
			&attributesJSON,
			&instance.ParentInstanceID,
			&instance.BranchID,
			&instance.IsParallelBranch,
			&instance.AuditDetails.CreatedBy,
			&instance.AuditDetails.CreatedTime,
			&instance.AuditDetails.ModifiedBy,
			&instance.AuditDetails.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

// GetSLABreachedInstances retrieves process instances that have breached SLA thresholds
func (r *processInstanceRepository) GetSLABreachedInstances(ctx context.Context, tenantID, processID, stateCode string, stateSlaMinutes, processSlaMinutes *int) ([]*models.ProcessInstance, error) {
	// Base query to get latest instances for each entity in the specified state
	query := `
		SELECT pi.id, pi.tenant_id, pi.process_id, pi.entity_id, pi.action, pi.status, 
			   pi.current_state_id, pi.documents, pi.assignees, pi.attributes, pi.comment,
			   pi.state_sla, pi.process_sla, pi.parent_instance_id, pi.branch_id, pi.is_parallel_branch,
			   pi.escalated, pi.created_by, pi.created_at, pi.modified_by, pi.modified_at
		FROM process_instances pi
		INNER JOIN (
			SELECT entity_id, MAX(created_at) as latest_time
			FROM process_instances 
			WHERE tenant_id = $1 AND process_id = $2 AND current_state_id = $3
			GROUP BY entity_id
		) latest ON pi.entity_id = latest.entity_id AND pi.created_at = latest.latest_time
		WHERE pi.tenant_id = $1 AND pi.process_id = $2 AND pi.current_state_id = $3`

	args := []interface{}{tenantID, processID, stateCode}
	argIndex := 3

	// Add SLA breach conditions
	currentTimeMillis := time.Now().UnixMilli()

	if stateSlaMinutes != nil && *stateSlaMinutes > 0 {
		argIndex++
		query += fmt.Sprintf(" AND (%d - pi.created_at) > ($%d * 60 * 1000)", currentTimeMillis, argIndex)
		args = append(args, *stateSlaMinutes)
	}

	if processSlaMinutes != nil && *processSlaMinutes > 0 {
		argIndex++
		query += fmt.Sprintf(" AND (%d - pi.created_at) > ($%d * 60 * 1000)", currentTimeMillis, argIndex)
		args = append(args, *processSlaMinutes)
	}

	query += " ORDER BY pi.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query SLA breached instances: %w", err)
	}
	defer rows.Close()

	var instances []*models.ProcessInstance
	for rows.Next() {
		var instance models.ProcessInstance
		var documentsJSON, assigneesJSON, attributesJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessID,
			&instance.EntityID,
			&instance.Action,
			&instance.Status,
			&instance.CurrentState,
			&documentsJSON,
			&assigneesJSON,
			&attributesJSON,
			&instance.Comment,
			&instance.StateSLA,
			&instance.ProcessSLA,
			&instance.ParentInstanceID,
			&instance.BranchID,
			&instance.IsParallelBranch,
			&instance.Escalated,
			&instance.AuditDetails.CreatedBy,
			&instance.AuditDetails.CreatedTime,
			&instance.AuditDetails.ModifiedBy,
			&instance.AuditDetails.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

// GetEscalatedInstances retrieves process instances that have been auto-escalated
// Following the Java service pattern - this searches for instances with escalated = true
func (r *processInstanceRepository) GetEscalatedInstances(ctx context.Context, tenantID, processID string, limit, offset int) ([]*models.ProcessInstance, error) {
	// Base query to get latest instances per entity that have been escalated
	// Following the exact Java service pattern: WHERE escalated = true
	query := `
		SELECT pi.id, pi.tenant_id, pi.process_id, pi.entity_id, pi.action, pi.status, 
			   pi.current_state_id, pi.documents, pi.assignees, pi.attributes, pi.comment,
			   pi.state_sla, pi.process_sla, pi.parent_instance_id, pi.branch_id, pi.is_parallel_branch,
			   pi.escalated, pi.created_by, pi.created_at, pi.modified_by, pi.modified_at
		FROM (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY entity_id ORDER BY created_at DESC) as rank_number
			FROM process_instances 
			WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 1

	if processID != "" {
		argIndex++
		query += fmt.Sprintf(" AND process_id = $%d", argIndex)
		args = append(args, processID)
	}

	query += `
		) pi 
		WHERE pi.rank_number = 1 
		AND pi.escalated = true`

	query += " ORDER BY pi.created_at DESC"

	if limit > 0 {
		argIndex++
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	if offset > 0 {
		argIndex++
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query escalated instances: %w", err)
	}
	defer rows.Close()

	var instances []*models.ProcessInstance
	for rows.Next() {
		var instance models.ProcessInstance
		var documentsJSON, assigneesJSON, attributesJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessID,
			&instance.EntityID,
			&instance.Action,
			&instance.Status,
			&instance.CurrentState,
			&documentsJSON,
			&assigneesJSON,
			&attributesJSON,
			&instance.Comment,
			&instance.StateSLA,
			&instance.ProcessSLA,
			&instance.ParentInstanceID,
			&instance.BranchID,
			&instance.IsParallelBranch,
			&instance.Escalated,
			&instance.AuditDetails.CreatedBy,
			&instance.AuditDetails.CreatedTime,
			&instance.AuditDetails.ModifiedBy,
			&instance.AuditDetails.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(documentsJSON) > 0 {
			json.Unmarshal(documentsJSON, &instance.Documents)
		}
		if len(assigneesJSON) > 0 {
			json.Unmarshal(assigneesJSON, &instance.Assignees)
		}
		if len(attributesJSON) > 0 {
			json.Unmarshal(attributesJSON, &instance.Attributes)
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}
