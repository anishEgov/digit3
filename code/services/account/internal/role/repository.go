package role

import (
	"account/internal/db"
	"account/internal/models"
	"context"
	"encoding/json"
)

func CreateRole(ctx context.Context, accountID string, r *models.Role) error {
	permissionsJSON, err := json.Marshal(r.Permissions)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
    INSERT INTO roles (
      role_id, account_id, name, permissions,
      created_by, created_on, modified_by, modified_on
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  `,
		r.RoleID, accountID, r.Name, permissionsJSON,
		r.CreatedBy, r.CreatedOn, r.ModifiedBy, r.ModifiedOn,
	)
	return err
}

func GetRoles(ctx context.Context, accountID string) ([]models.Role, error) {
	rows, err := db.Pool.Query(ctx, `
    SELECT role_id, name, permissions,
           created_by, created_on, modified_by, modified_on
    FROM roles WHERE account_id = $1
  `, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		var permissionsJSON []byte
		if err := rows.Scan(&r.RoleID, &r.Name, &permissionsJSON,
			&r.CreatedBy, &r.CreatedOn, &r.ModifiedBy, &r.ModifiedOn); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permissionsJSON, &r.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	return roles, nil
}
