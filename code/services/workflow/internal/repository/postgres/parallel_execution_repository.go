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

type parallelExecutionRepository struct {
	db *sqlx.DB
}

func NewParallelExecutionRepository(db *sqlx.DB) repository.ParallelExecutionRepository {
	return &parallelExecutionRepository{db: db}
}

func (r *parallelExecutionRepository) CreateParallelExecution(ctx context.Context, execution *models.ParallelExecution) error {
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	// Marshal JSON fields
	activeBranchesJSON, err := json.Marshal(execution.ActiveBranches)
	if err != nil {
		return fmt.Errorf("error marshaling active branches: %w", err)
	}

	completedBranchesJSON, err := json.Marshal(execution.CompletedBranches)
	if err != nil {
		return fmt.Errorf("error marshaling completed branches: %w", err)
	}

	query := `INSERT INTO parallel_executions (id, tenant_id, entity_id, process_id, parallel_state_id, join_state_id, active_branches, completed_branches, status, created_by, created_at, modified_by, modified_at)
	VALUES (:id, :tenant_id, :entity_id, :process_id, :parallel_state_id, :join_state_id, :active_branches, :completed_branches, :status, :created_by, :created_at, :modified_by, :modified_at)`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                 execution.ID,
		"tenant_id":          execution.TenantID,
		"entity_id":          execution.EntityID,
		"process_id":         execution.ProcessID,
		"parallel_state_id":  execution.ParallelStateID,
		"join_state_id":      execution.JoinStateID,
		"active_branches":    activeBranchesJSON,
		"completed_branches": completedBranchesJSON,
		"status":             execution.Status,
		"created_by":         execution.AuditDetail.CreatedBy,
		"created_at":         execution.AuditDetail.CreatedTime,
		"modified_by":        execution.AuditDetail.ModifiedBy,
		"modified_at":        execution.AuditDetail.ModifiedTime,
	})
	return err
}

func (r *parallelExecutionRepository) GetParallelExecution(ctx context.Context, tenantID, entityID, processID, parallelStateID string) (*models.ParallelExecution, error) {
	var execution models.ParallelExecution
	query := `SELECT id, tenant_id, entity_id, process_id, parallel_state_id, join_state_id, active_branches, completed_branches, status, created_by, created_at, modified_by, modified_at
		FROM parallel_executions WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND parallel_state_id = $4`

	var activeBranchesJSON, completedBranchesJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, entityID, processID, parallelStateID).Scan(
		&execution.ID,
		&execution.TenantID,
		&execution.EntityID,
		&execution.ProcessID,
		&execution.ParallelStateID,
		&execution.JoinStateID,
		&activeBranchesJSON,
		&completedBranchesJSON,
		&execution.Status,
		&execution.AuditDetail.CreatedBy,
		&execution.AuditDetail.CreatedTime,
		&execution.AuditDetail.ModifiedBy,
		&execution.AuditDetail.ModifiedTime,
	)

	if err == nil {
		// Unmarshal JSON fields
		if len(activeBranchesJSON) > 0 {
			json.Unmarshal(activeBranchesJSON, &execution.ActiveBranches)
		}
		if len(completedBranchesJSON) > 0 {
			json.Unmarshal(completedBranchesJSON, &execution.CompletedBranches)
		}
	}

	return &execution, err
}

func (r *parallelExecutionRepository) UpdateParallelExecution(ctx context.Context, execution *models.ParallelExecution) error {
	// Marshal JSON fields
	activeBranchesJSON, err := json.Marshal(execution.ActiveBranches)
	if err != nil {
		return fmt.Errorf("error marshaling active branches: %w", err)
	}

	completedBranchesJSON, err := json.Marshal(execution.CompletedBranches)
	if err != nil {
		return fmt.Errorf("error marshaling completed branches: %w", err)
	}

	query := `UPDATE parallel_executions SET 
		active_branches = :active_branches,
		completed_branches = :completed_branches,
		status = :status,
		modified_by = :modified_by,
		modified_at = :modified_at
		WHERE tenant_id = :tenant_id AND id = :id`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                 execution.ID,
		"tenant_id":          execution.TenantID,
		"active_branches":    activeBranchesJSON,
		"completed_branches": completedBranchesJSON,
		"status":             execution.Status,
		"modified_by":        execution.AuditDetail.ModifiedBy,
		"modified_at":        execution.AuditDetail.ModifiedTime,
	})
	return err
}

func (r *parallelExecutionRepository) MarkBranchCompleted(ctx context.Context, tenantID, entityID, processID, branchID string) error {
	query := `UPDATE parallel_executions SET 
		completed_branches = completed_branches || $5::jsonb,
		modified_at = $6
		WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND status = 'ACTIVE'
		AND NOT (completed_branches ? $4)`

	branchJSON, err := json.Marshal(branchID)
	if err != nil {
		return fmt.Errorf("error marshaling branch ID: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, tenantID, entityID, processID, branchID, branchJSON, time.Now().UnixMilli())
	return err
}

func (r *parallelExecutionRepository) GetActiveParallelExecutions(ctx context.Context, tenantID, entityID, processID string) ([]*models.ParallelExecution, error) {
	query := `SELECT id, tenant_id, entity_id, process_id, parallel_state_id, join_state_id, active_branches, completed_branches, status, created_by, created_at, modified_by, modified_at
		FROM parallel_executions WHERE tenant_id = $1 AND entity_id = $2 AND process_id = $3 AND status = 'ACTIVE'`

	rows, err := r.db.QueryContext(ctx, query, tenantID, entityID, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*models.ParallelExecution
	for rows.Next() {
		var execution models.ParallelExecution
		var activeBranchesJSON, completedBranchesJSON []byte

		err := rows.Scan(
			&execution.ID,
			&execution.TenantID,
			&execution.EntityID,
			&execution.ProcessID,
			&execution.ParallelStateID,
			&execution.JoinStateID,
			&activeBranchesJSON,
			&completedBranchesJSON,
			&execution.Status,
			&execution.AuditDetail.CreatedBy,
			&execution.AuditDetail.CreatedTime,
			&execution.AuditDetail.ModifiedBy,
			&execution.AuditDetail.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(activeBranchesJSON) > 0 {
			json.Unmarshal(activeBranchesJSON, &execution.ActiveBranches)
		}
		if len(completedBranchesJSON) > 0 {
			json.Unmarshal(completedBranchesJSON, &execution.CompletedBranches)
		}

		executions = append(executions, &execution)
	}

	return executions, rows.Err()
}
