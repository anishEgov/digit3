package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	ID   string
	SQL  string
	File string
}

// LoadMigrations loads all SQL migration files from the migrations directory
func LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	// Use relative path from working directory
	migrationsDir := "./internal/migrations"
	
	// Check if directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Fallback to runtime caller method
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return nil, fmt.Errorf("failed to get current file path")
		}
		migrationsDir = filepath.Dir(filename)
	}

	// Read all files in the migrations directory
	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process .sql files
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", path, err)
			}

			// Extract migration ID from filename (e.g., "001_create_tables.sql" -> "001")
			baseName := filepath.Base(path)
			parts := strings.Split(baseName, "_")
			if len(parts) == 0 {
				return fmt.Errorf("invalid migration filename: %s", baseName)
			}
			migrationID := parts[0]

			migrations = append(migrations, Migration{
				ID:   migrationID,
				SQL:  string(content),
				File: baseName,
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Sort migrations by ID to ensure proper order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// RunMigrations executes all migrations on the database
func RunMigrations(db *sql.DB) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, migration := range migrations {
		fmt.Printf("Running migration: %s\n", migration.File)
		
		// Split SQL by semicolon to handle multiple statements
		statements := strings.Split(migration.SQL, ";")
		
		for _, statement := range statements {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			
			_, err := db.Exec(statement)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.File, err)
			}
		}
		
		fmt.Printf("Completed migration: %s\n", migration.File)
	}

	return nil
}
