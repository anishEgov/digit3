package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	Validators ValidatorConfig
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

type ValidatorConfig struct {
	EnabledValidators []string
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

	// Parse VALIDATORS environment variable
	validatorsEnv := getEnv("VALIDATORS", "roles") // Default to "roles" only
	enabledValidators := parseValidators(validatorsEnv)

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
		},
		Validators: ValidatorConfig{
			EnabledValidators: enabledValidators,
		},
	}, nil
}

func parseValidators(validatorsStr string) []string {
	if validatorsStr == "" {
		return []string{"roles"} // Default fallback
	}

	validators := strings.Split(validatorsStr, ",")
	// Trim whitespace from each validator name
	for i, validator := range validators {
		validators[i] = strings.TrimSpace(validator)
	}

	return validators
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
