package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"github.com/jmoiron/sqlx"
)

type escalationConfigRepository struct {
	db *sqlx.DB
}

// NewEscalationConfigRepository creates a new instance of EscalationConfigRepository.
func NewEscalationConfigRepository(db *sqlx.DB) repository.EscalationConfigRepository {
	return &escalationConfigRepository{db: db}
}

// CreateEscalationConfig creates a new escalation configuration.
func (r *escalationConfigRepository) CreateEscalationConfig(ctx context.Context, config *models.EscalationConfig) (*models.EscalationConfig, error) {
	query := `
		INSERT INTO escalation_configs (
			tenant_id, process_id, state_code, escalation_action, 
			state_sla_minutes, process_sla_minutes,
			created_by, created_at, modified_by, modified_at
		) VALUES (
			:tenant_id, :process_id, :state_code, :escalation_action,
			:state_sla_minutes, :process_sla_minutes,
			:created_by, :created_at, :modified_by, :modified_at
		) RETURNING id`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare escalation config creation query: %w", err)
	}
	defer stmt.Close()

	var id string
	if err := stmt.GetContext(ctx, &id, config); err != nil {
		return nil, fmt.Errorf("failed to create escalation config: %w", err)
	}

	config.ID = id
	return config, nil
}

// GetEscalationConfigByID retrieves an escalation configuration by ID.
func (r *escalationConfigRepository) GetEscalationConfigByID(ctx context.Context, tenantID, id string) (*models.EscalationConfig, error) {
	query := `
		SELECT id, tenant_id, process_id, state_code, escalation_action,
			   state_sla_minutes, process_sla_minutes,
			   created_by, created_at, modified_by, modified_at
		FROM escalation_configs 
		WHERE tenant_id = $1 AND id = $2`

	var config models.EscalationConfig
	if err := r.db.GetContext(ctx, &config, query, tenantID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("escalation config not found")
		}
		return nil, fmt.Errorf("failed to get escalation config: %w", err)
	}

	return &config, nil
}

// GetEscalationConfigsByProcessID retrieves escalation configurations for a process.
func (r *escalationConfigRepository) GetEscalationConfigsByProcessID(ctx context.Context, tenantID, processID string, stateCode string, isActive *bool) ([]*models.EscalationConfig, error) {
	query := `
		SELECT id, tenant_id, process_id, state_code, escalation_action,
			   state_sla_minutes, process_sla_minutes,
			   created_by, created_at, modified_by, modified_at
		FROM escalation_configs 
		WHERE tenant_id = $1 AND process_id = $2`

	args := []interface{}{tenantID, processID}
	argIndex := 2

	if stateCode != "" {
		argIndex++
		query += fmt.Sprintf(" AND state_code = $%d", argIndex)
		args = append(args, stateCode)
	}

	// Note: isActive parameter is ignored since we removed the column
	// We keep the parameter for API compatibility but don't use it

	query += " ORDER BY state_code, escalation_action"

	var configs []*models.EscalationConfig
	if err := r.db.SelectContext(ctx, &configs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get escalation configs: %w", err)
	}

	return configs, nil
}

// UpdateEscalationConfig updates an existing escalation configuration.
func (r *escalationConfigRepository) UpdateEscalationConfig(ctx context.Context, config *models.EscalationConfig) (*models.EscalationConfig, error) {
	// First, fetch the existing config to preserve unchanged fields
	existing, err := r.GetEscalationConfigByID(ctx, config.TenantID, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing escalation config: %w", err)
	}

	// Merge non-empty fields from request into existing config
	if config.StateCode != "" {
		existing.StateCode = config.StateCode
	}
	if config.EscalationAction != "" {
		existing.EscalationAction = config.EscalationAction
	}
	if config.StateSlaMinutes != nil {
		existing.StateSlaMinutes = config.StateSlaMinutes
	}
	if config.ProcessSlaMinutes != nil {
		existing.ProcessSlaMinutes = config.ProcessSlaMinutes
	}
	// Update audit details
	existing.AuditDetail.ModifiedBy = config.AuditDetail.ModifiedBy
	existing.AuditDetail.ModifiedTime = config.AuditDetail.ModifiedTime

	query := `
		UPDATE escalation_configs SET
			state_code = :state_code,
			escalation_action = :escalation_action,
			state_sla_minutes = :state_sla_minutes,
			process_sla_minutes = :process_sla_minutes,
			modified_by = :modified_by,
			modified_at = :modified_at
		WHERE tenant_id = :tenant_id AND id = :id`

	if _, err := r.db.NamedExecContext(ctx, query, existing); err != nil {
		return nil, fmt.Errorf("failed to update escalation config: %w", err)
	}

	return existing, nil
}

// DeleteEscalationConfig deletes an escalation configuration.
func (r *escalationConfigRepository) DeleteEscalationConfig(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM escalation_configs WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.ExecContext(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete escalation config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("escalation config not found")
	}

	return nil
}

// GetActiveEscalationConfigs retrieves all escalation configurations for a tenant.
// Note: Method name kept for compatibility, but returns all configs since isActive was removed.
func (r *escalationConfigRepository) GetActiveEscalationConfigs(ctx context.Context, tenantID string) ([]*models.EscalationConfig, error) {
	query := `
		SELECT id, tenant_id, process_id, state_code, escalation_action,
			   state_sla_minutes, process_sla_minutes,
			   created_by, created_at, modified_by, modified_at
		FROM escalation_configs 
		WHERE tenant_id = $1
		ORDER BY process_id, state_code`

	var configs []*models.EscalationConfig
	if err := r.db.SelectContext(ctx, &configs, query, tenantID); err != nil {
		return nil, fmt.Errorf("failed to get escalation configs: %w", err)
	}

	return configs, nil
}
