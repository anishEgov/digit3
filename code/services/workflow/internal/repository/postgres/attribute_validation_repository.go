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

type attributeValidationRepository struct {
	db *sqlx.DB
}

// NewAttributeValidationRepository creates a new instance of AttributeValidationRepository.
func NewAttributeValidationRepository(db *sqlx.DB) repository.AttributeValidationRepository {
	return &attributeValidationRepository{db: db}
}

// CreateAttributeValidation inserts a new attribute validation record.
func (r *attributeValidationRepository) CreateAttributeValidation(ctx context.Context, validation *models.AttributeValidation) error {
	validation.ID = uuid.New().String()

	// Set default audit values if not provided
	if validation.AuditDetail.CreatedBy == "" {
		validation.AuditDetail.CreatedBy = "system"
	}
	if validation.AuditDetail.ModifiedBy == "" {
		validation.AuditDetail.ModifiedBy = "system"
	}

	now := time.Now().UnixMilli()
	validation.AuditDetail.CreatedTime = now
	validation.AuditDetail.ModifiedTime = now

	attributesJSON, err := json.Marshal(validation.Attributes)
	if err != nil {
		return err
	}

	query := `INSERT INTO attribute_validations (id, tenant_id, attributes, assignee_check, created_by, created_at, modified_by, modified_at)
              VALUES (:id, :tenant_id, :attributes, :assignee_check, :created_by, :created_at, :modified_by, :modified_at)`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             validation.ID,
		"tenant_id":      validation.TenantID,
		"attributes":     attributesJSON,
		"assignee_check": validation.AssigneeCheck,
		"created_by":     validation.AuditDetail.CreatedBy,
		"created_at":     validation.AuditDetail.CreatedTime,
		"modified_by":    validation.AuditDetail.ModifiedBy,
		"modified_at":    validation.AuditDetail.ModifiedTime,
	})
	return err
}

// GetAttributeValidationByID retrieves a single attribute validation by its ID.
func (r *attributeValidationRepository) GetAttributeValidationByID(ctx context.Context, tenantID, id string) (*models.AttributeValidation, error) {
	type validationRow struct {
		ID            string `db:"id"`
		TenantID      string `db:"tenant_id"`
		Attributes    []byte `db:"attributes"`
		AssigneeCheck bool   `db:"assignee_check"`
		CreatedBy     string `db:"created_by"`
		CreatedAt     int64  `db:"created_at"`
		ModifiedBy    string `db:"modified_by"`
		ModifiedAt    int64  `db:"modified_at"`
	}

	var row validationRow
	query := `SELECT id, tenant_id, attributes, assignee_check, created_by, created_at, modified_by, modified_at 
              FROM attribute_validations WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &row, query, tenantID, id)
	if err != nil {
		return nil, err
	}

	var attributes map[string][]string
	if len(row.Attributes) > 0 {
		err := json.Unmarshal(row.Attributes, &attributes)
		if err != nil {
			return nil, err
		}
	}

	validation := &models.AttributeValidation{
		ID:            row.ID,
		TenantID:      row.TenantID,
		Attributes:    attributes,
		AssigneeCheck: row.AssigneeCheck,
		AuditDetail: models.AuditDetail{
			CreatedBy:    row.CreatedBy,
			CreatedTime:  row.CreatedAt,
			ModifiedBy:   row.ModifiedBy,
			ModifiedTime: row.ModifiedAt,
		},
	}
	return validation, nil
}
