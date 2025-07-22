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

// CreateState inserts a new state record into the database.
func (r *stateRepository) CreateState(ctx context.Context, state *models.State) error {
	state.ID = uuid.New().String()
	now := time.Now().UnixMilli()
	state.AuditDetail.CreatedTime = now
	state.AuditDetail.ModifiedTime = now

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
	type stateRow struct {
		ID           string  `db:"id"`
		TenantID     string  `db:"tenant_id"`
		ProcessID    string  `db:"process_id"`
		Code         string  `db:"code"`
		Name         string  `db:"name"`
		Description  *string `db:"description"`
		SLA          *int64  `db:"sla"`
		IsInitial    bool    `db:"is_initial"`
		IsParallel   bool    `db:"is_parallel"`
		IsJoin       bool    `db:"is_join"`
		BranchStates []byte  `db:"branch_states"`
		CreatedBy    string  `db:"created_by"`
		CreatedAt    int64   `db:"created_at"`
		ModifiedBy   string  `db:"modified_by"`
		ModifiedAt   int64   `db:"modified_at"`
	}

	var rows []stateRow
	query := `SELECT id, tenant_id, process_id, code, name, description, sla, is_initial, is_parallel, is_join, branch_states, created_by, created_at, modified_by, modified_at
	          FROM states WHERE tenant_id = $1 AND process_id = $2`
	err := r.db.SelectContext(ctx, &rows, query, tenantID, processID)
	if err != nil {
		return nil, err
	}

	states := make([]*models.State, 0, len(rows))
	for _, row := range rows {
		var branchStates []string
		if len(row.BranchStates) > 0 {
			err := json.Unmarshal(row.BranchStates, &branchStates)
			if err != nil {
				return nil, err
			}
		}

		state := &models.State{
			ID:           row.ID,
			TenantID:     row.TenantID,
			ProcessID:    row.ProcessID,
			Code:         row.Code,
			Name:         row.Name,
			Description:  row.Description,
			SLA:          row.SLA,
			IsInitial:    row.IsInitial,
			IsParallel:   row.IsParallel,
			IsJoin:       row.IsJoin,
			BranchStates: branchStates,
			AuditDetail: models.AuditDetail{
				CreatedBy:    row.CreatedBy,
				CreatedTime:  row.CreatedAt,
				ModifiedBy:   row.ModifiedBy,
				ModifiedTime: row.ModifiedAt,
			},
		}
		states = append(states, state)
	}
	return states, nil
}

// GetStateByID retrieves a single state by its ID.
func (r *stateRepository) GetStateByID(ctx context.Context, tenantID, id string) (*models.State, error) {
	type stateRow struct {
		ID           string  `db:"id"`
		TenantID     string  `db:"tenant_id"`
		ProcessID    string  `db:"process_id"`
		Code         string  `db:"code"`
		Name         string  `db:"name"`
		Description  *string `db:"description"`
		SLA          *int64  `db:"sla"`
		IsInitial    bool    `db:"is_initial"`
		IsParallel   bool    `db:"is_parallel"`
		IsJoin       bool    `db:"is_join"`
		BranchStates []byte  `db:"branch_states"`
		CreatedBy    string  `db:"created_by"`
		CreatedAt    int64   `db:"created_at"`
		ModifiedBy   string  `db:"modified_by"`
		ModifiedAt   int64   `db:"modified_at"`
	}

	var row stateRow
	query := `SELECT id, tenant_id, process_id, code, name, description, sla, is_initial, is_parallel, is_join, branch_states, created_by, created_at, modified_by, modified_at
	          FROM states WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &row, query, tenantID, id)
	if err != nil {
		return nil, err
	}

	var branchStates []string
	if len(row.BranchStates) > 0 {
		err := json.Unmarshal(row.BranchStates, &branchStates)
		if err != nil {
			return nil, err
		}
	}

	state := &models.State{
		ID:           row.ID,
		TenantID:     row.TenantID,
		ProcessID:    row.ProcessID,
		Code:         row.Code,
		Name:         row.Name,
		Description:  row.Description,
		SLA:          row.SLA,
		IsInitial:    row.IsInitial,
		IsParallel:   row.IsParallel,
		IsJoin:       row.IsJoin,
		BranchStates: branchStates,
		AuditDetail: models.AuditDetail{
			CreatedBy:    row.CreatedBy,
			CreatedTime:  row.CreatedAt,
			ModifiedBy:   row.ModifiedBy,
			ModifiedTime: row.ModifiedAt,
		},
	}
	return state, nil
}

// UpdateState updates an existing state record.
func (r *stateRepository) UpdateState(ctx context.Context, state *models.State) error {
	state.AuditDetail.ModifiedTime = time.Now().UnixMilli()

	branchStatesJSON, err := json.Marshal(state.BranchStates)
	if err != nil {
		return err
	}

	query := `UPDATE states SET
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

// DeleteState removes a state record from the database.
func (r *stateRepository) DeleteState(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM states WHERE tenant_id = $1 AND id = $2"
	_, err := r.db.ExecContext(ctx, query, tenantID, id)
	return err
}
