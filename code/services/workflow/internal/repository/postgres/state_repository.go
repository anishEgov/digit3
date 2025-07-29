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

type stateRepository struct {
	db *sqlx.DB
}

// NewStateRepository creates a new instance of StateRepository.
func NewStateRepository(db *sqlx.DB) repository.StateRepository {
	return &stateRepository{db: db}
}

// CreateState inserts a new state record.
func (r *stateRepository) CreateState(ctx context.Context, state *models.State) error {
	state.ID = uuid.New().String()
	// Audit details should be set by handlers, only set time if not already set
	now := time.Now().UnixMilli()
	if state.AuditDetail.CreatedTime == 0 {
		state.AuditDetail.CreatedTime = now
	}
	if state.AuditDetail.ModifiedTime == 0 {
		state.AuditDetail.ModifiedTime = now
	}

	branchStatesJSON, err := json.Marshal(state.BranchStates)
	if err != nil {
		return err
	}

	query := `INSERT INTO states (id, tenant_id, process_id, code, name, description, sla, is_initial, is_parallel, is_join, branch_states, created_by, created_at, modified_by, modified_at)
              VALUES (:id, :tenant_id, :process_id, :code, :name, :description, :sla, :is_initial, :is_parallel, :is_join, :branch_states, :created_by, :created_at, :modified_by, :modified_at)`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":            state.ID,
		"tenant_id":     state.TenantID,
		"process_id":    state.ProcessID,
		"code":          state.Code,
		"name":          state.Name,
		"description":   state.Description,
		"sla":           state.SLA,
		"is_initial":    state.IsInitial,
		"is_parallel":   state.IsParallel,
		"is_join":       state.IsJoin,
		"branch_states": branchStatesJSON,
		"created_by":    state.AuditDetail.CreatedBy,
		"created_at":    state.AuditDetail.CreatedTime,
		"modified_by":   state.AuditDetail.ModifiedBy,
		"modified_at":   state.AuditDetail.ModifiedTime,
	})
	return err
}

// GetStatesByProcessID retrieves all states for a given process.
func (r *stateRepository) GetStatesByProcessID(ctx context.Context, tenantID, processID string) ([]*models.State, error) {
	query := `SELECT * FROM states WHERE tenant_id = $1 AND process_id = $2 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, tenantID, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*models.State
	for rows.Next() {
		var state models.State
		var branchStatesJSON []byte

		err := rows.Scan(
			&state.ID,
			&state.TenantID,
			&state.ProcessID,
			&state.Code,
			&state.Name,
			&state.Description,
			&state.SLA,
			&state.IsInitial,
			&state.IsParallel,
			&state.IsJoin,
			&branchStatesJSON,
			&state.AuditDetail.CreatedBy,
			&state.AuditDetail.CreatedTime,
			&state.AuditDetail.ModifiedBy,
			&state.AuditDetail.ModifiedTime,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal branch_states JSONB field
		if len(branchStatesJSON) > 0 {
			json.Unmarshal(branchStatesJSON, &state.BranchStates)
		}

		states = append(states, &state)
	}

	return states, rows.Err()
}

// GetStateByID retrieves a single state by its ID.
func (r *stateRepository) GetStateByID(ctx context.Context, tenantID, id string) (*models.State, error) {
	query := `SELECT * FROM states WHERE tenant_id = $1 AND id = $2`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)

	var state models.State
	var branchStatesJSON []byte

	err := row.Scan(
		&state.ID,
		&state.TenantID,
		&state.ProcessID,
		&state.Code,
		&state.Name,
		&state.Description,
		&state.SLA,
		&state.IsInitial,
		&state.IsParallel,
		&state.IsJoin,
		&branchStatesJSON,
		&state.AuditDetail.CreatedBy,
		&state.AuditDetail.CreatedTime,
		&state.AuditDetail.ModifiedBy,
		&state.AuditDetail.ModifiedTime,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal branch_states JSONB field
	if len(branchStatesJSON) > 0 {
		json.Unmarshal(branchStatesJSON, &state.BranchStates)
	}

	return &state, nil
}

// UpdateState updates an existing state record.
func (r *stateRepository) UpdateState(ctx context.Context, state *models.State) error {
	now := time.Now().UnixMilli()
	state.AuditDetail.ModifiedTime = now

	branchStatesJSON, err := json.Marshal(state.BranchStates)
	if err != nil {
		return err
	}

	query := `UPDATE states 
			  SET code = :code,
				  name = :name,
				  description = :description,
				  sla = :sla,
				  is_initial = :is_initial,
				  is_parallel = :is_parallel,
				  is_join = :is_join,
				  branch_states = :branch_states,
				  modified_by = :modified_by,
				  modified_at = :modified_at
			  WHERE tenant_id = :tenant_id AND id = :id`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":            state.ID,
		"tenant_id":     state.TenantID,
		"code":          state.Code,
		"name":          state.Name,
		"description":   state.Description,
		"sla":           state.SLA,
		"is_initial":    state.IsInitial,
		"is_parallel":   state.IsParallel,
		"is_join":       state.IsJoin,
		"branch_states": branchStatesJSON,
		"modified_by":   state.AuditDetail.ModifiedBy,
		"modified_at":   state.AuditDetail.ModifiedTime,
	})
	return err
}

// DeleteState deletes a state record.
// GetStateByCodeAndProcess finds a state by code within a specific process.
func (r *stateRepository) GetStateByCodeAndProcess(ctx context.Context, tenantID, processID, code string) (*models.State, error) {
	query := `SELECT id, tenant_id, process_id, code, name, description, sla, is_initial, is_parallel, is_join, branch_states, created_by, created_at, modified_by, modified_at 
              FROM states WHERE tenant_id = $1 AND process_id = $2 AND code = $3`

	var state models.State
	var branchStatesBytes []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, processID, code).Scan(
		&state.ID, &state.TenantID, &state.ProcessID, &state.Code, &state.Name, &state.Description,
		&state.SLA, &state.IsInitial, &state.IsParallel, &state.IsJoin, &branchStatesBytes,
		&state.AuditDetail.CreatedBy, &state.AuditDetail.CreatedTime,
		&state.AuditDetail.ModifiedBy, &state.AuditDetail.ModifiedTime,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal branch_states JSONB
	if branchStatesBytes != nil {
		if err := json.Unmarshal(branchStatesBytes, &state.BranchStates); err != nil {
			return nil, err
		}
	}

	return &state, nil
}

func (r *stateRepository) DeleteState(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM states WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, id)
	return err
}
