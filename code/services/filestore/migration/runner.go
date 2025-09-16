package migration

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Runner handles database migrations
type Runner struct {
	db     *sqlx.DB
	config *Config
}

// NewRunner creates a new migration runner
func NewRunner(db *sqlx.DB, config *Config) *Runner {
	return &Runner{
		db:     db,
		config: config,
	}
}

// Run executes pending migrations with override capability
func (r *Runner) Run(ctx context.Context) error {
	if !r.config.Enabled {
		log.Println("Migrations disabled, skipping...")
		return nil
	}

	log.Println("Starting database migrations...")

	// Create migration tracking table
	if err := r.createMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Acquire lock
	if err := r.acquireLock(); err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer r.releaseLock()

	// Load migration files
	migrations, err := r.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := r.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get pending migrations (including changed ones for override)
	pending := r.getPendingMigrations(migrations, applied)

	if len(pending) == 0 {
		log.Println("No pending migrations found")
		return nil
	}

	log.Printf("Found %d pending migrations", len(pending))

	// Execute pending migrations
	for _, migration := range pending {
		if err := r.executeMigration(migration); err != nil {
			// Log error but continue with other migrations for override capability
			log.Printf("Failed to execute migration %s: %v (continuing with other migrations)", migration.Version, err)
			continue
		}
		log.Printf("Applied migration %s: %s", migration.Version, migration.Name)
	}

	log.Println("Migration process completed")
	return nil
}

// createMigrationTable creates the migration tracking table
func (r *Runner) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS filestore_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(32) NOT NULL
		)`

	_, err := r.db.Exec(query)
	return err
}

// acquireLock acquires an advisory lock to prevent concurrent migrations
func (r *Runner) acquireLock() error {
	_, err := r.db.Exec("SELECT pg_advisory_lock(987654321)")
	return err
}

// releaseLock releases the advisory lock
func (r *Runner) releaseLock() {
	r.db.Exec("SELECT pg_advisory_unlock(987654321)")
}

// loadMigrations loads migration files from the filesystem
func (r *Runner) loadMigrations() ([]*Migration, error) {
	var migrations []*Migration

	err := filepath.WalkDir(r.config.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		filename := d.Name()
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			return nil // Skip files that don't match the pattern
		}

		version := parts[0]
		nameWithSuffix := strings.TrimSuffix(parts[1], ".sql")
		name := strings.TrimSuffix(nameWithSuffix, ".up") // Remove .up suffix if present

		// Read file content for checksum
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		checksum := fmt.Sprintf("%x", md5.Sum(content))

		migrations = append(migrations, &Migration{
			Version:  version,
			Name:     name,
			Filepath: path,
			Checksum: checksum,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		vi, _ := strconv.Atoi(migrations[i].Version)
		vj, _ := strconv.Atoi(migrations[j].Version)
		return vi < vj
	})

	return migrations, nil
}

// getAppliedMigrations retrieves applied migrations from the database
func (r *Runner) getAppliedMigrations() (map[string]*Migration, error) {
	query := "SELECT version, name, applied_at, checksum FROM filestore_migrations"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]*Migration)
	for rows.Next() {
		var migration Migration
		var appliedAt time.Time

		err := rows.Scan(&migration.Version, &migration.Name, &appliedAt, &migration.Checksum)
		if err != nil {
			return nil, err
		}

		migration.AppliedAt = &appliedAt
		migration.IsApplied = true
		applied[migration.Version] = &migration
	}

	return applied, rows.Err()
}

// getPendingMigrations returns migrations that haven't been applied yet or have changed (for override)
func (r *Runner) getPendingMigrations(migrations []*Migration, applied map[string]*Migration) []*Migration {
	var pending []*Migration

	for _, migration := range migrations {
		if appliedMigration, exists := applied[migration.Version]; exists {
			// Check if checksum matches
			if appliedMigration.Checksum != migration.Checksum {
				log.Printf("Migration %s has changed since it was applied - will re-run for override", migration.Version)
				pending = append(pending, migration)
			}
			continue
		}
		pending = append(pending, migration)
	}

	return pending
}

// executeMigration executes a single migration with override capability
func (r *Runner) executeMigration(migration *Migration) error {
	// Read migration file
	content, err := os.ReadFile(migration.Filepath)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL (continue on errors for override capability)
	sqlContent := string(content)
	statements := strings.Split(sqlContent, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.Exec(stmt); err != nil {
			// Log the error but continue for override capability
			log.Printf("Warning: Error executing statement in migration %s: %v", migration.Version, err)
		}
	}

	// Record/Update migration as applied
	query := `
		INSERT INTO filestore_migrations (version, name, applied_at, checksum)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (version) DO UPDATE SET
			name = EXCLUDED.name,
			applied_at = EXCLUDED.applied_at,
			checksum = EXCLUDED.checksum`

	if _, err := tx.Exec(query, migration.Version, migration.Name, time.Now(), migration.Checksum); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}
