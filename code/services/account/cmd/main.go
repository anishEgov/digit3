package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"account/internal/account"
	"account/internal/db"
	"account/internal/role"
	"account/internal/user"
)

func main() {
	ctx := context.Background()
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbPool, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	db.Init(dbPool)
	if err := db.InitDB(dbPool); err != nil {
		log.Fatalf("DB initialization failed: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v3/accounts", func(r chi.Router) {
		r.Post("/", account.CreateAccountHandler)
		r.Get("/{id}", account.GetAccountHandler)
		r.Post("/{id}/status", account.UpdateAccountStatusHandler)
		r.Post("/{id}/administrator", account.UpdateAdministratorHandler)
		r.Post("/{id}/users", user.AddUserHandler)
		r.Get("/{id}/users", user.ListUsersHandler)
		r.Post("/{id}/roles", role.CreateRoleHandler)
		r.Get("/{id}/roles", role.ListRolesHandler)
	})

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
