package role

import (
	"account/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	var req struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	role := models.Role{
		RoleID:      uuid.NewString(),
		Name:        req.Name,
		Permissions: req.Permissions,
		CreatedBy:   "system",
		CreatedOn:   time.Now().UTC(),
		ModifiedBy:  "system",
		ModifiedOn:  time.Now().UTC(),
	}

	if err := CreateRole(r.Context(), accountID, &role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

func ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	roles, err := GetRoles(r.Context(), accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}
