package postgres

import (
	"context"
	"fmt"
	"time"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type processInstanceRepository struct {
	db *gorm.DB
}

func NewProcessInstanceRepository(db *gorm.DB) repository.ProcessInstanceRepository {
	return &processInstanceRepository{db: db}
}

func generateUUID() string {
	return uuid.New().String()
}

func (r *processInstanceRepository) CreateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error {
	if instance.ID == "" {
		instance.ID = uuid.New().String()
	}

	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *processInstanceRepository) GetProcessInstanceByID(ctx context.Context, tenantID, id string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *processInstanceRepository) GetProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_id = ? AND process_id = ?", tenantID, entityID, processID).
		First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *processInstanceRepository) GetLatestProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error) {
	var instance models.ProcessInstance
	fmt.Printf("🔍 REPO DEBUG: Getting latest instance for entityID=%s, processID=%s\n", entityID, processID)

	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_id = ? AND process_id = ? AND is_parallel_branch = false", tenantID, entityID, processID).
		Order("created_at DESC").
		First(&instance).Error

	if err == nil {
		fmt.Printf("✅ REPO DEBUG: Found latest instance ID=%s, currentState=%s, action=%s, createdAt=%d\n",
			instance.ID, instance.CurrentState, instance.Action, instance.AuditDetails.CreatedTime)
	} else {
		fmt.Printf("❌ REPO DEBUG: Error getting latest instance: %v\n", err)
	}

	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *processInstanceRepository) GetProcessInstancesByEntityID(ctx context.Context, tenantID, entityID, processID string, history bool) ([]*models.ProcessInstance, error) {
	var instances []*models.ProcessInstance

	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_id = ? AND process_id = ?", tenantID, entityID, processID)

	if history {
		// Return all records ordered by created_at (oldest first for chronological order)
		query = query.Order("created_at ASC")
	} else {
		// Return only the latest record
		query = query.Order("created_at DESC").Limit(1)
	}

	err := query.Find(&instances).Error
	return instances, err
}

func (r *processInstanceRepository) UpdateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error {
	// Update audit details
	now := time.Now().UnixMilli()
	instance.AuditDetails.ModifiedTime = now
	if instance.AuditDetails.ModifiedBy == "" {
		instance.AuditDetails.ModifiedBy = "system"
	}

	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", instance.TenantID, instance.ID).
		Updates(instance).Error
}

// GetActiveParallelInstances returns all active parallel branch instances for an entity
func (r *processInstanceRepository) GetActiveParallelInstances(ctx context.Context, tenantID, entityID, processID string) ([]*models.ProcessInstance, error) {
	var instances []*models.ProcessInstance
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_id = ? AND process_id = ? AND is_parallel_branch = true AND status != ?",
			tenantID, entityID, processID, "COMPLETED").
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

// GetInstancesByBranch returns instances for a specific parallel branch
func (r *processInstanceRepository) GetInstancesByBranch(ctx context.Context, tenantID, entityID, processID, branchID string) ([]*models.ProcessInstance, error) {
	var instances []*models.ProcessInstance
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_id = ? AND process_id = ? AND branch_id = ?",
			tenantID, entityID, processID, branchID).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

// GetSLABreachedInstances retrieves process instances that have breached SLA thresholds
func (r *processInstanceRepository) GetSLABreachedInstances(ctx context.Context, tenantID, processID, stateCode string, stateSlaMinutes, processSlaMinutes *int) ([]*models.ProcessInstance, error) {
	// Use raw SQL for complex subquery with window functions - GORM doesn't handle this well
	currentTimeMillis := time.Now().UnixMilli()

	query := `
		SELECT pi.id, pi.tenant_id, pi.process_id, pi.entity_id, pi.action, pi.status, 
			   pi.current_state_id, pi.documents, pi.assignees, pi.attributes, pi.comment,
			   pi.state_sla, pi.process_sla, pi.parent_instance_id, pi.branch_id, pi.is_parallel_branch,
			   pi.escalated, pi.created_by, pi.created_at, pi.modified_by, pi.modified_at
		FROM process_instances pi
		INNER JOIN (
			SELECT entity_id, MAX(created_at) as latest_time
			FROM process_instances 
			WHERE tenant_id = ? AND process_id = ? AND current_state_id = ?
			GROUP BY entity_id
		) latest ON pi.entity_id = latest.entity_id AND pi.created_at = latest.latest_time
		WHERE pi.tenant_id = ? AND pi.process_id = ? AND pi.current_state_id = ?`

	args := []interface{}{tenantID, processID, stateCode, tenantID, processID, stateCode}

	// Add SLA breach conditions
	if stateSlaMinutes != nil && *stateSlaMinutes > 0 {
		query += fmt.Sprintf(" AND (%d - pi.created_at) > (? * 60 * 1000)", currentTimeMillis)
		args = append(args, *stateSlaMinutes)
	}

	if processSlaMinutes != nil && *processSlaMinutes > 0 {
		query += fmt.Sprintf(" AND (%d - pi.created_at) > (? * 60 * 1000)", currentTimeMillis)
		args = append(args, *processSlaMinutes)
	}

	query += " ORDER BY pi.created_at DESC"

	var instances []*models.ProcessInstance
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query SLA breached instances: %w", err)
	}

	return instances, nil
}

// GetEscalatedInstances retrieves process instances that have been auto-escalated
// Following the Java service pattern - this searches for instances with escalated = true
func (r *processInstanceRepository) GetEscalatedInstances(ctx context.Context, tenantID, processID string, limit, offset int) ([]*models.ProcessInstance, error) {
	// Use raw SQL for complex window function query - GORM doesn't handle this well
	query := `
		SELECT pi.id, pi.tenant_id, pi.process_id, pi.entity_id, pi.action, pi.status, 
			   pi.current_state_id, pi.documents, pi.assignees, pi.attributes, pi.comment,
			   pi.state_sla, pi.process_sla, pi.parent_instance_id, pi.branch_id, pi.is_parallel_branch,
			   pi.escalated, pi.created_by, pi.created_at, pi.modified_by, pi.modified_at
		FROM (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY entity_id ORDER BY created_at DESC) as rank_number
			FROM process_instances 
			WHERE tenant_id = ?`

	args := []interface{}{tenantID}

	if processID != "" {
		query += " AND process_id = ?"
		args = append(args, processID)
	}

	query += `
		) pi 
		WHERE pi.rank_number = 1 
		AND pi.escalated = true
		ORDER BY pi.created_at DESC`

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	var instances []*models.ProcessInstance
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query escalated instances: %w", err)
	}

	return instances, nil
}
