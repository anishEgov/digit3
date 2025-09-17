package db

import (
	"account/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	cfg := config.LoadConfig()
	// Build DSN from config fields
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
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	DB = db
	createTables(db)
	fmt.Println("Database connection established and tables checked/created.")
}

func createTables(db *sql.DB) {
	tenantTable := `CREATE TABLE IF NOT EXISTS tenant_v1 (
		id VARCHAR(128) PRIMARY KEY,
		code VARCHAR(255) UNIQUE NOT NULL,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(512) NOT NULL,
		additionalAttributes JSONB,
		isActive BOOLEAN NOT NULL,
		tenantId VARCHAR(64),
		createdBy VARCHAR(64),
		lastModifiedBy VARCHAR(64),
		createdTime BIGINT,
		lastModifiedTime BIGINT
	);`

	tenantConfigTable := `CREATE TABLE IF NOT EXISTS tenant_config_v1 (
		id VARCHAR(128) PRIMARY KEY,
		code VARCHAR(255) NOT NULL,
		defaultLoginType VARCHAR(50),
		otpLength VARCHAR(50),
		name VARCHAR(255),
		enableUserBasedLogin BOOLEAN,
		additionalAttributes JSONB,
		isActive BOOLEAN NOT NULL,
		languages TEXT[],
		createdBy VARCHAR(64),
		lastModifiedBy VARCHAR(64),
		createdTime BIGINT,
		lastModifiedTime BIGINT
	);`

	tenantDocumentTable := `CREATE TABLE IF NOT EXISTS tenant_documents_v1 (
		id VARCHAR(128) PRIMARY KEY DEFAULT gen_random_uuid(),
		tenantConfigId VARCHAR(128) REFERENCES tenant_config_v1(id) ON DELETE CASCADE,
		tenantId VARCHAR(255) NOT NULL,
		type VARCHAR(255),
		fileStoreId VARCHAR(255),
		url VARCHAR(512),
		isActive BOOLEAN NOT NULL,
		createdBy VARCHAR(64),
		lastModifiedBy VARCHAR(64),
		createdTime BIGINT,
		lastModifiedTime BIGINT
	);`

	_, err := db.Exec(tenantTable)
	if err != nil {
		log.Fatalf("failed to create tenant_v1 table: %v", err)
	}
	_, err = db.Exec(tenantConfigTable)
	if err != nil {
		log.Fatalf("failed to create tenant_config_v1 table: %v", err)
	}
	_, err = db.Exec(tenantDocumentTable)
	if err != nil {
		log.Fatalf("failed to create tenant_documents_v1 table: %v", err)
	}
}
