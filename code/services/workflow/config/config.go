package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Migration MigrationConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type MigrationConfig struct {
	RunMigrations bool
	MigrationPath string
	Timeout       time.Duration
}

func LoadConfig() (*Config, error) {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	// Parse migration settings
	runMigrations, _ := strconv.ParseBool(getEnv("RUN_MIGRATIONS", "true"))
	migrationTimeout, _ := time.ParseDuration(getEnv("MIGRATION_TIMEOUT", "5m"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
		},
		Migration: MigrationConfig{
			RunMigrations: runMigrations,
			MigrationPath: getEnv("MIGRATION_PATH", "db/migration"),
			Timeout:       migrationTimeout,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
