// internal/account/handler.go
package account

import (
	"account/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string            `json:"name"`
		Domain        string            `json:"domain"`
		OIDCConfig    models.OIDCConfig `json:"oidc_config"`
		Administrator string            `json:"administrator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	account := models.Account{
		ID:            uuid.NewString(),
		Name:          req.Name,
		Domain:        req.Domain,
		Status:        "active",
		Administrator: req.Administrator,
		OIDCConfig:    req.OIDCConfig,
		CreatedBy:     req.Administrator,
		CreatedOn:     time.Now().UTC(),
		ModifiedBy:    req.Administrator,
		ModifiedOn:    time.Now().UTC(),
	}

	err := CreateAccount(r.Context(), &account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	account, err := GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func UpdateAccountStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Status != "active" && req.Status != "closed" {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}
	if err := UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func UpdateAdministratorHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Administrator string `json:"administrator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Administrator == "" {
		http.Error(w, "Administrator ID required", http.StatusBadRequest)
		return
	}
	if err := UpdateAdministrator(r.Context(), id, req.Administrator); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
