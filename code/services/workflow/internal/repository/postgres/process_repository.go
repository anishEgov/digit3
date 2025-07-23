package postgres

import (
	"context"
	"time"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type processRepository struct {
	db *sqlx.DB
}

// NewProcessRepository creates a new instance of ProcessRepository.
func NewProcessRepository(db *sqlx.DB) repository.ProcessRepository {
	return &processRepository{db: db}
}

// CreateProcess inserts a new process record into the database.
func (r *processRepository) CreateProcess(ctx context.Context, process *models.Process) error {
	process.ID = uuid.New().String()
	now := time.Now().UnixMilli()
	process.AuditDetail.CreatedTime = now
	process.AuditDetail.ModifiedTime = now

	query := `INSERT INTO processes (id, tenant_id, name, code, description, version, sla, created_by, created_at, modified_by, modified_at)
              VALUES (:id, :tenant_id, :name, :code, :description, :version, :sla, :created_by, :created_at, :modified_by, :modified_at)`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          process.ID,
		"tenant_id":   process.TenantID,
		"name":        process.Name,
		"code":        process.Code,
		"description": process.Description,
		"version":     process.Version,
		"sla":         process.SLA,
		"created_by":  process.AuditDetail.CreatedBy,
		"created_at":  process.AuditDetail.CreatedTime,
		"modified_by": process.AuditDetail.ModifiedBy,
		"modified_at": process.AuditDetail.ModifiedTime,
	})

	return err
}

// GetProcessByID retrieves a single process by its ID.
func (r *processRepository) GetProcessByID(ctx context.Context, tenantID, id string) (*models.Process, error) {
	var process models.Process
	query := `SELECT id, tenant_id, name, code, description, version, sla, 
	                 created_by, created_at, modified_by, modified_at 
	          FROM processes WHERE tenant_id = $1 AND id = $2`
	err := r.db.GetContext(ctx, &process, query, tenantID, id)
	return &process, err
}

// GetProcessByCode retrieves a single process by its code.
func (r *processRepository) GetProcessByCode(ctx context.Context, tenantID, code string) (*models.Process, error) {
	var process models.Process
	query := `SELECT id, tenant_id, name, code, description, version, sla, 
	                 created_by, created_at, modified_by, modified_at 
	          FROM processes WHERE tenant_id = $1 AND code = $2`
	err := r.db.GetContext(ctx, &process, query, tenantID, code)
	return &process, err
}

// GetProcesses retrieves a list of processes based on filter criteria.
func (r *processRepository) GetProcesses(ctx context.Context, tenantID string, ids []string, names []string) ([]*models.Process, error) {
	query := `SELECT id, tenant_id, name, code, description, version, sla, 
	                 created_by, created_at, modified_by, modified_at 
	          FROM processes WHERE tenant_id = ?`
	args := []interface{}{tenantID}

	if len(ids) > 0 {
		query += " AND id IN (?)"
		args = append(args, ids)
	}
	if len(names) > 0 {
		query += " AND name IN (?)"
		args = append(args, names)
	}

	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var processes []*models.Process
	err = r.db.SelectContext(ctx, &processes, query, args...)
	return processes, err
}

// UpdateProcess updates an existing process record in the database.
func (r *processRepository) UpdateProcess(ctx context.Context, process *models.Process) error {
	process.AuditDetail.ModifiedTime = time.Now().UnixMilli()

	query := `UPDATE processes SET 
                  name = :name, 
                  code = :code, 
                  description = :description, 
                  version = :version, 
                  sla = :sla, 
                  modified_by = :modified_by, 
                  modified_at = :modified_at
              WHERE tenant_id = :tenant_id AND id = :id`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          process.ID,
		"tenant_id":   process.TenantID,
		"name":        process.Name,
		"code":        process.Code,
		"description": process.Description,
		"version":     process.Version,
		"sla":         process.SLA,
		"modified_by": process.AuditDetail.ModifiedBy,
		"modified_at": process.AuditDetail.ModifiedTime,
	})
	return err
}

// DeleteProcess removes a process record from the database.
func (r *processRepository) DeleteProcess(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM processes WHERE tenant_id = $1 AND id = $2"
	_, err := r.db.ExecContext(ctx, query, tenantID, id)
	return err
}
