package postgres

import (
	"context"
	"fmt"
	"time"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type actionRepository struct {
	db                      *sqlx.DB
	attributeValidationRepo repository.AttributeValidationRepository
}

// NewActionRepository creates a new instance of ActionRepository.
func NewActionRepository(db *sqlx.DB, attributeValidationRepo repository.AttributeValidationRepository) repository.ActionRepository {
	return &actionRepository{
		db:                      db,
		attributeValidationRepo: attributeValidationRepo,
	}
}

// CreateAction creates a new action in the database.
func (r *actionRepository) CreateAction(ctx context.Context, action *models.Action) (*models.Action, error) {
	// Generate UUID if not provided
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	now := time.Now().UnixMilli()
	createdBy := action.AuditDetail.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}

	// Handle AttributeValidation if provided
	var attributeValidationID *string
	if action.AttributeValidation != nil {
		action.AttributeValidation.TenantID = action.TenantID
		err := r.attributeValidationRepo.CreateAttributeValidation(ctx, action.AttributeValidation)
		if err != nil {
			return nil, fmt.Errorf("failed to create attribute validation: %w", err)
		}
		attributeValidationID = &action.AttributeValidation.ID
	}

	query := `INSERT INTO actions (id, tenant_id, name, label, current_state_id, next_state_id, attribute_validation_id, created_by, created_at, modified_by, modified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, modified_at`

	var returnedID string
	var createdAt, modifiedAt int64
	err := r.db.QueryRowContext(ctx, query,
		action.ID,             // $1
		action.TenantID,       // $2
		action.Name,           // $3
		action.Label,          // $4
		action.CurrentState,   // $5
		action.NextState,      // $6
		attributeValidationID, // $7
		createdBy,             // $8
		now,                   // $9
		createdBy,             // $10
		now,                   // $11
	).Scan(&returnedID, &createdAt, &modifiedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	action.ID = returnedID
	action.AuditDetail.CreatedBy = createdBy
	action.AuditDetail.CreatedTime = createdAt
	action.AuditDetail.ModifiedBy = createdBy
	action.AuditDetail.ModifiedTime = modifiedAt

	return action, nil
}

// GetActionsByStateID retrieves all actions for a given state.
func (r *actionRepository) GetActionsByStateID(ctx context.Context, tenantID, stateID string) ([]*models.Action, error) {
	type actionRow struct {
		ID                    string  `db:"id"`
		TenantID              string  `db:"tenant_id"`
		Name                  string  `db:"name"`
		Label                 *string `db:"label"`
		CurrentStateID        string  `db:"current_state_id"`
		NextStateID           string  `db:"next_state_id"`
		AttributeValidationID *string `db:"attribute_validation_id"`
		CreatedBy             string  `db:"created_by"`
		CreatedAt             int64   `db:"created_at"`
		ModifiedBy            string  `db:"modified_by"`
		ModifiedAt            int64   `db:"modified_at"`
	}

	var rows []actionRow
	query := `SELECT id, tenant_id, name, label, current_state_id, next_state_id, attribute_validation_id, created_by, created_at, modified_by, modified_at FROM actions WHERE tenant_id = $1 AND current_state_id = $2`
	err := r.db.SelectContext(ctx, &rows, query, tenantID, stateID)
	if err != nil {
		return nil, err
	}

	actions := make([]*models.Action, 0, len(rows))
	for _, row := range rows {
		action := &models.Action{
			ID:                    row.ID,
			TenantID:              row.TenantID,
			Name:                  row.Name,
			Label:                 row.Label,
			CurrentState:          row.CurrentStateID,
			NextState:             row.NextStateID,
			AttributeValidationID: row.AttributeValidationID,
			AuditDetail: models.AuditDetail{
				CreatedBy:    row.CreatedBy,
				CreatedTime:  row.CreatedAt,
				ModifiedBy:   row.ModifiedBy,
				ModifiedTime: row.ModifiedAt,
			},
		}

		// Load AttributeValidation if ID exists
		if row.AttributeValidationID != nil {
			attributeValidation, err := r.attributeValidationRepo.GetAttributeValidationByID(ctx, tenantID, *row.AttributeValidationID)
			if err == nil {
				action.AttributeValidation = attributeValidation
			}
		}

		actions = append(actions, action)
	}
	return actions, nil
}

// GetActionByID retrieves a single action by its ID.
func (r *actionRepository) GetActionByID(ctx context.Context, tenantID, id string) (*models.Action, error) {
	type actionRow struct {
		ID                    string  `db:"id"`
		TenantID              string  `db:"tenant_id"`
		Name                  string  `db:"name"`
		Label                 *string `db:"label"`
		CurrentStateID        string  `db:"current_state_id"`
		NextStateID           string  `db:"next_state_id"`
		AttributeValidationID *string `db:"attribute_validation_id"`
		CreatedBy             string  `db:"created_by"`
		CreatedAt             int64   `db:"created_at"`
		ModifiedBy            string  `db:"modified_by"`
		ModifiedAt            int64   `db:"modified_at"`
	}

	var row actionRow
	query := `SELECT id, tenant_id, name, label, current_state_id, next_state_id, attribute_validation_id, created_by, created_at, modified_by, modified_at FROM actions WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &row, query, tenantID, id)
	if err != nil {
		return nil, err
	}

	action := &models.Action{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		Name:                  row.Name,
		Label:                 row.Label,
		CurrentState:          row.CurrentStateID,
		NextState:             row.NextStateID,
		AttributeValidationID: row.AttributeValidationID,
		AuditDetail: models.AuditDetail{
			CreatedBy:    row.CreatedBy,
			CreatedTime:  row.CreatedAt,
			ModifiedBy:   row.ModifiedBy,
			ModifiedTime: row.ModifiedAt,
		},
	}

	// Load AttributeValidation if ID exists
	if row.AttributeValidationID != nil {
		attributeValidation, err := r.attributeValidationRepo.GetAttributeValidationByID(ctx, tenantID, *row.AttributeValidationID)
		if err == nil {
			action.AttributeValidation = attributeValidation
		}
	}

	return action, nil
}

// UpdateAction updates an existing action in the database.
func (r *actionRepository) UpdateAction(ctx context.Context, action *models.Action) error {
	action.AuditDetail.ModifiedTime = time.Now().UnixMilli()

	query := `UPDATE actions SET
				  name = :name,
				  label = :label,
				  next_state_id = :next_state_id,
				  attribute_validation_id = :attribute_validation_id,
				  modified_by = :modified_by,
				  modified_at = :modified_at
			  WHERE tenant_id = :tenant_id AND id = :id`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                      action.ID,
		"tenant_id":               action.TenantID,
		"name":                    action.Name,
		"label":                   action.Label,
		"next_state_id":           action.NextState,
		"attribute_validation_id": action.AttributeValidationID,
		"modified_by":             action.AuditDetail.ModifiedBy,
		"modified_at":             action.AuditDetail.ModifiedTime,
	})
	return err
}

// DeleteAction removes an action record from the database.
func (r *actionRepository) DeleteAction(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM actions WHERE tenant_id = $1 AND id = $2"
	_, err := r.db.ExecContext(ctx, query, tenantID, id)
	return err
}
