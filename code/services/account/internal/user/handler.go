package user

import (
	"account/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	var req struct {
		Name     string   `json:"name"`
		Email    string   `json:"email"`
		Phone    string   `json:"phone"`
		UniqueID string   `json:"unique_id"`
		Roles    []string `json:"roles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user := models.User{
		UserID:     uuid.NewString(),
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		UniqueID:   req.UniqueID,
		Roles:      req.Roles,
		CreatedBy:  "system",
		CreatedOn:  time.Now().UTC(),
		ModifiedBy: "system",
		ModifiedOn: time.Now().UTC(),
	}

	if err := CreateUser(r.Context(), accountID, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	users, err := GetUsers(r.Context(), accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
