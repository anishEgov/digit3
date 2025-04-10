package account

import (
	"account/internal/db"
	"account/internal/models"
	"context"
	"encoding/json"
	"errors"
)

func CreateAccount(ctx context.Context, a *models.Account) error {
	oidcJSON, err := json.Marshal(a.OIDCConfig)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO accounts (
			id, name, domain, status, administrator,
			oidc_config, created_by, created_on, modified_by, modified_on
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		a.ID, a.Name, a.Domain, a.Status, a.Administrator,
		oidcJSON, a.CreatedBy, a.CreatedOn, a.ModifiedBy, a.ModifiedOn)

	return err
}

func GetAccount(ctx context.Context, id string) (*models.Account, error) {
	row := db.Pool.QueryRow(ctx, `
		SELECT id, name, domain, status, administrator,
		       oidc_config, created_by, created_on, modified_by, modified_on
		FROM accounts WHERE id=$1
	`, id)

	var a models.Account
	var oidcJSON []byte
	if err := row.Scan(&a.ID, &a.Name, &a.Domain, &a.Status, &a.Administrator,
		&oidcJSON, &a.CreatedBy, &a.CreatedOn, &a.ModifiedBy, &a.ModifiedOn); err != nil {
		return nil, errors.New("account not found")
	}
	if err := json.Unmarshal(oidcJSON, &a.OIDCConfig); err != nil {
		return nil, err
	}

	return &a, nil
}

func UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE accounts SET status=$1, modified_on=NOW() WHERE id=$2`, status, id)
	return err
}

func UpdateAdministrator(ctx context.Context, id string, administrator string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE accounts SET administrator=$1, modified_on=NOW() WHERE id=$2`, administrator, id)
	return err
}
