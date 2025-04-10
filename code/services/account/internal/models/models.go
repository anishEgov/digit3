// internal/models/models.go
package models

import "time"

type OIDCConfig struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"client_id"`
}

type Account struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Domain        string     `json:"domain"`
	Status        string     `json:"status"`
	Administrator string     `json:"administrator"`
	OIDCConfig    OIDCConfig `json:"oidc_config"`
	CreatedBy     string     `json:"created_by"`
	CreatedOn     time.Time  `json:"created_on"`
	ModifiedBy    string     `json:"modified_by"`
	ModifiedOn    time.Time  `json:"modified_on"`
}

type User struct {
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	UniqueID   string    `json:"unique_id"`
	Roles      []string  `json:"roles"`
	CreatedBy  string    `json:"created_by"`
	CreatedOn  time.Time `json:"created_on"`
	ModifiedBy string    `json:"modified_by"`
	ModifiedOn time.Time `json:"modified_on"`
}

type Role struct {
	RoleID      string    `json:"role_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	CreatedBy   string    `json:"created_by"`
	CreatedOn   time.Time `json:"created_on"`
	ModifiedBy  string    `json:"modified_by"`
	ModifiedOn  time.Time `json:"modified_on"`
}
