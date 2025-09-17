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

	"gorm.io/gorm"
)

// Runner handles database migrations
type Runner struct {
	db     *gorm.DB
	config *Config
}

// NewRunner creates a new migration runner
func NewRunner(db *gorm.DB, config *Config) *Runner {
	return &Runner{
		db:     db,
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

	log.Println("All migrations completed successfully")
	return nil
}

// createMigrationTable creates the migration tracking table
func (r *Runner) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS url_shortener_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(32) NOT NULL
		)`
	return r.db.Exec(query).Error
}

// acquireLock acquires an advisory lock to prevent concurrent migrations
func (r *Runner) acquireLock() error {
	return r.db.Exec("SELECT pg_advisory_lock(123456789)").Error
}

// releaseLock releases the advisory lock
func (r *Runner) releaseLock() {
	r.db.Exec("SELECT pg_advisory_unlock(123456789)")
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
		name := strings.TrimSuffix(parts[1], ".sql")

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
	var applied []Migration
	if err := r.db.Raw("SELECT version, name, applied_at, checksum FROM url_shortener_migrations").Scan(&applied).Error; err != nil {
		return nil, err
	}

	appliedMap := make(map[string]*Migration)
	for i := range applied {
		m := applied[i]
		m.IsApplied = true
		appliedMap[m.Version] = &m
	}

	return appliedMap, nil
}

// getPendingMigrations returns migrations that haven't been applied yet
func (r *Runner) getPendingMigrations(migrations []*Migration, applied map[string]*Migration) []*Migration {
	var pending []*Migration

	for _, migration := range migrations {
		if appliedMigration, exists := applied[migration.Version]; exists {
			// Check if checksum matches
			if appliedMigration.Checksum != migration.Checksum {
				log.Printf("Warning: Migration %s has changed since it was applied (checksum mismatch)", migration.Version)
			}
			continue
		}
		pending = append(pending, migration)
	}

	return pending
}

// executeMigration executes a single migration
func (r *Runner) executeMigration(migration *Migration) error {
	// Read migration file
	content, err := os.ReadFile(migration.Filepath)
	if err != nil {
		return err
	}

	// Start transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Execute migration SQL
	if err := tx.Exec(string(content)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	if err := tx.Exec(
		"INSERT INTO url_shortener_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
		migration.Version, migration.Name, time.Now(), migration.Checksum,
	).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	return tx.Commit().Error
}
