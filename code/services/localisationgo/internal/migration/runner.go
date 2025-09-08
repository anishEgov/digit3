package migration

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Runner handles database migrations
type Runner struct {
	db     *sqlx.DB
	config *Config
}

// NewRunner creates a new migration runner
func NewRunner(db *sql.DB, config *Config) *Runner {
	sqlxDB := sqlx.NewDb(db, "postgres")
	return &Runner{
		db:     sqlxDB,
		config: config,
	}
}

// Run executes pending migrations
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

	// Get pending migrations
	pending := r.getPendingMigrations(migrations, applied)

	if len(pending) == 0 {
		log.Println("No pending migrations found")
		return nil
	}

	log.Printf("Found %d pending migrations", len(pending))

	// Execute pending migrations
	for _, migration := range pending {
		if err := r.executeMigration(migration); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}
		log.Printf("Applied migration %s: %s", migration.Version, migration.Name)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createMigrationTable creates the migration tracking table
func (r *Runner) createMigrationTable() error {
	// Check if old golang-migrate schema_migrations table exists
	var hasNameColumn bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'schema_migrations' 
			AND column_name = 'name'
		)
	`).Scan(&hasNameColumn)

	if err != nil {
		return fmt.Errorf("failed to check schema_migrations structure: %w", err)
	}

	if !hasNameColumn {
		// Old golang-migrate table exists, need to migrate it
		log.Println("Detected old golang-migrate schema_migrations table, migrating to new format...")

		// Rename old table
		if _, err := r.db.Exec("ALTER TABLE schema_migrations RENAME TO schema_migrations_old"); err != nil {
			return fmt.Errorf("failed to rename old schema_migrations table: %w", err)
		}

		// Create new table
		query := `
			CREATE TABLE schema_migrations (
				version VARCHAR(255) PRIMARY KEY,
				name VARCHAR(500) NOT NULL,
				checksum VARCHAR(32) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`
		if _, err := r.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create new schema_migrations table: %w", err)
		}

		// Check what migrations were applied based on database state
		if err := r.migrateFromOldSchema(); err != nil {
			return fmt.Errorf("failed to migrate from old schema: %w", err)
		}

		// Drop old table
		if _, err := r.db.Exec("DROP TABLE schema_migrations_old"); err != nil {
			log.Printf("Warning: failed to drop old schema_migrations table: %v", err)
		}

		log.Println("Successfully migrated schema_migrations table to new format")
	} else {
		// Table already exists with correct structure or doesn't exist
		query := `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version VARCHAR(255) PRIMARY KEY,
				name VARCHAR(500) NOT NULL,
				checksum VARCHAR(32) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`
		if _, err := r.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// acquireLock acquires a database lock for migrations
func (r *Runner) acquireLock() error {
	// Use PostgreSQL advisory lock
	var acquired bool
	err := r.db.QueryRow("SELECT pg_try_advisory_lock(123456789)").Scan(&acquired)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("could not acquire migration lock")
	}
	return nil
}

// releaseLock releases the database lock
func (r *Runner) releaseLock() {
	r.db.Exec("SELECT pg_advisory_unlock(123456789)")
}

// loadMigrations loads all migration files from the configured path
func (r *Runner) loadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(r.config.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Extract version from filename (e.g., "001_create_table.sql" -> "001")
		filename := d.Name()
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		// Calculate checksum
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		checksum := fmt.Sprintf("%x", md5.Sum(content))

		migrations = append(migrations, Migration{
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
func (r *Runner) getAppliedMigrations() (map[string]Migration, error) {
	applied := make(map[string]Migration)

	rows, err := r.db.Query("SELECT version, name, checksum, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.Name, &m.Checksum, &m.AppliedAt); err != nil {
			return nil, err
		}
		m.IsApplied = true
		applied[m.Version] = m
	}

	return applied, rows.Err()
}

// getPendingMigrations returns migrations that haven't been applied
func (r *Runner) getPendingMigrations(all []Migration, applied map[string]Migration) []Migration {
	var pending []Migration

	for _, migration := range all {
		if appliedMigration, exists := applied[migration.Version]; exists {
			// Check if checksum matches
			if appliedMigration.Checksum != migration.Checksum {
				log.Printf("WARNING: Migration %s checksum mismatch. Applied: %s, Current: %s",
					migration.Version, appliedMigration.Checksum, migration.Checksum)
			}
			continue
		}
		pending = append(pending, migration)
	}

	return pending
}

// executeMigration executes a single migration
func (r *Runner) executeMigration(migration Migration) error {
	// Read migration file
	content, err := os.ReadFile(migration.Filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)",
		migration.Version, migration.Name, migration.Checksum,
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// migrateFromOldSchema migrates from golang-migrate schema to new format
func (r *Runner) migrateFromOldSchema() error {
	// Check database state to determine what migrations should be marked as applied
	var tableExists bool

	// Check if localisation table exists
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'localisation'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check localisation table: %w", err)
	}

	if !tableExists {
		// No localisation table, no migrations to mark as applied
		return nil
	}

	log.Println("Found existing localisation table, determining applied migrations...")

	// Migration 1: Create table - if table exists, this is applied
	if err := r.markMigrationApplied("001", "create_localisation_table", "auto-migrated"); err != nil {
		return err
	}

	// Migration 2: Alter user ID type - check if created_by is VARCHAR
	var dataType string
	err = r.db.QueryRow(`
		SELECT data_type FROM information_schema.columns 
		WHERE table_name = 'localisation' AND column_name = 'created_by'
	`).Scan(&dataType)
	if err == nil && dataType == "character varying" {
		if err := r.markMigrationApplied("002", "alter_user_id_type", "auto-migrated"); err != nil {
			return err
		}
	}

	// Migration 3: Add UUID - check if uuid column exists
	var uuidExists bool
	err = r.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_name = 'localisation' AND column_name = 'uuid'
		)
	`).Scan(&uuidExists)
	if err == nil && uuidExists {
		if err := r.markMigrationApplied("003", "add_uuid_to_localisation", "auto-migrated"); err != nil {
			return err
		}
	}

	return nil
}

// markMigrationApplied marks a migration as applied with the given checksum
func (r *Runner) markMigrationApplied(version, name, checksum string) error {
	_, err := r.db.Exec(
		"INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)",
		version, name, checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to mark migration %s as applied: %w", version, err)
	}
	log.Printf("Marked migration %s (%s) as applied", version, name)
	return nil
}
