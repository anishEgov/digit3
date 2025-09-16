package internal

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"idgen/internal/config"
	"idgen/internal/migrations"

	_ "github.com/lib/pq"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	// Build DSN from config
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	
	// Retry ping for up to 10 seconds
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("Waiting for DB to be ready...")
		time.Sleep(1 * time.Second)
	}

	// Run database migrations
	if err := migrations.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connection established and migrations completed.")
	return db, nil
}
