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

// CreateAction inserts a new action record.
func (r *actionRepository) CreateAction(ctx context.Context, action *models.Action) error {
	action.ID = uuid.New().String()

	// Validate required UUID fields
	if action.CurrentState == "" {
		return fmt.Errorf("current_state_id cannot be empty")
	}
	if action.NextState == "" {
		return fmt.Errorf("next_state_id cannot be empty")
	}

	// Set default audit values only if not already provided by handlers
	if action.AuditDetail.CreatedBy == "" {
		action.AuditDetail.CreatedBy = "system"
	}
	if action.AuditDetail.ModifiedBy == "" {
		action.AuditDetail.ModifiedBy = "system"
	}

	now := time.Now().UnixMilli()
	if action.AuditDetail.CreatedTime == 0 {
		action.AuditDetail.CreatedTime = now
	}
	if action.AuditDetail.ModifiedTime == 0 {
		action.AuditDetail.ModifiedTime = now
	}

	rolesJSON, err := json.Marshal(action.Roles)
	if err != nil {
		return err
	}

	// Handle AttributeValidation if provided
	if action.AttributeValidation != nil {
		action.AttributeValidation.TenantID = action.TenantID
		err := r.attributeValidationRepo.CreateAttributeValidation(ctx, action.AttributeValidation)
		if err != nil {
			return err
		}
		action.AttributeValidationID = &action.AttributeValidation.ID
	}

	query := `INSERT INTO actions (id, tenant_id, name, label, current_state_id, next_state_id, roles, attribute_validation_id, created_by, created_at, modified_by, modified_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// Handle nil pointers properly for optional fields
	var attributeValidationID interface{}
	if action.AttributeValidationID != nil {
		attributeValidationID = *action.AttributeValidationID
	} else {
		attributeValidationID = nil
	}

	var label interface{}
	if action.Label != nil {
		label = *action.Label
	} else {
		label = nil
	}

	_, err = r.db.ExecContext(ctx, query,
		action.ID,                       // $1
		action.TenantID,                 // $2
		action.Name,                     // $3
		label,                           // $4
		action.CurrentState,             // $5
		action.NextState,                // $6
		rolesJSON,                       // $7
		attributeValidationID,           // $8
		action.AuditDetail.CreatedBy,    // $9
		action.AuditDetail.CreatedTime,  // $10
		action.AuditDetail.ModifiedBy,   // $11
		action.AuditDetail.ModifiedTime, // $12
	)
	return err
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
		Roles                 []byte  `db:"roles"`
		AttributeValidationID *string `db:"attribute_validation_id"`
		CreatedBy             string  `db:"created_by"`
		CreatedAt             int64   `db:"created_at"`
		ModifiedBy            string  `db:"modified_by"`
		ModifiedAt            int64   `db:"modified_at"`
	}

	var rows []actionRow
	query := `SELECT id, tenant_id, name, label, current_state_id, next_state_id, roles, attribute_validation_id, created_by, created_at, modified_by, modified_at FROM actions WHERE tenant_id = $1 AND current_state_id = $2`
	err := r.db.SelectContext(ctx, &rows, query, tenantID, stateID)
	if err != nil {
		return nil, err
	}

	actions := make([]*models.Action, 0, len(rows))
	for _, row := range rows {
		var roles []string
		if len(row.Roles) > 0 {
			err := json.Unmarshal(row.Roles, &roles)
			if err != nil {
				return nil, err
			}
		}

		action := &models.Action{
			ID:                    row.ID,
			TenantID:              row.TenantID,
			Name:                  row.Name,
			Label:                 row.Label,
			CurrentState:          row.CurrentStateID,
			NextState:             row.NextStateID,
			Roles:                 roles,
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
		Roles                 []byte  `db:"roles"`
		AttributeValidationID *string `db:"attribute_validation_id"`
		CreatedBy             string  `db:"created_by"`
		CreatedAt             int64   `db:"created_at"`
		ModifiedBy            string  `db:"modified_by"`
		ModifiedAt            int64   `db:"modified_at"`
	}

	var row actionRow
	query := `SELECT id, tenant_id, name, label, current_state_id, next_state_id, roles, attribute_validation_id, created_by, created_at, modified_by, modified_at FROM actions WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &row, query, tenantID, id)
	if err != nil {
		return nil, err
	}

	var roles []string
	if len(row.Roles) > 0 {
		err := json.Unmarshal(row.Roles, &roles)
		if err != nil {
			return nil, err
		}
	}

	action := &models.Action{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		Name:                  row.Name,
		Label:                 row.Label,
		CurrentState:          row.CurrentStateID,
		NextState:             row.NextStateID,
		Roles:                 roles,
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

// UpdateAction updates an existing action record.
func (r *actionRepository) UpdateAction(ctx context.Context, action *models.Action) error {
	action.AuditDetail.ModifiedTime = time.Now().UnixMilli()

	rolesJSON, err := json.Marshal(action.Roles)
	if err != nil {
		return err
	}

	query := `UPDATE actions SET
				  name = :name,
				  label = :label,
				  next_state_id = :next_state_id,
				  roles = :roles,
				  attribute_validation_id = :attribute_validation_id,
				  modified_by = :modified_by,
				  modified_at = :modified_at
			  WHERE tenant_id = :tenant_id AND id = :id`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                      action.ID,
		"tenant_id":               action.TenantID,
		"name":                    action.Name,
		"label":                   action.Label,
		"next_state_id":           action.NextState,
		"roles":                   rolesJSON,
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
