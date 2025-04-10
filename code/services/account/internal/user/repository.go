package user

import (
	"account/internal/db"
	"account/internal/models"
	"context"
	"encoding/json"
)

func CreateUser(ctx context.Context, accountID string, u *models.User) error {
	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (
			user_id, account_id, name, email, phone, unique_id,
			roles, created_by, created_on, modified_by, modified_on
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		u.UserID, accountID, u.Name, u.Email, u.Phone, u.UniqueID,
		rolesJSON, u.CreatedBy, u.CreatedOn, u.ModifiedBy, u.ModifiedOn,
	)
	return err
}

func GetUsers(ctx context.Context, accountID string) ([]models.User, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT user_id, name, email, phone, unique_id, roles,
		       created_by, created_on, modified_by, modified_on
		FROM users WHERE account_id = $1
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var rolesJSON []byte
		if err := rows.Scan(&u.UserID, &u.Name, &u.Email, &u.Phone, &u.UniqueID, &rolesJSON,
			&u.CreatedBy, &u.CreatedOn, &u.ModifiedBy, &u.ModifiedOn); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rolesJSON, &u.Roles); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
