package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Jaeger   JaegerConfig
}

type ServerConfig struct {
	RestPort string
	GrpcPort string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host string
	Port string
	DB   string
}

type JaegerConfig struct {
	SamplerParam string
	AgentHost    string
	AgentPort    string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			RestPort: getEnv("REST_PORT", "8088"),
			GrpcPort: getEnv("GRPC_PORT", "8089"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			DB:   getEnv("REDIS_DB", "0"),
		},
		Jaeger: JaegerConfig{
			SamplerParam: getEnv("JAEGER_SAMPLER_PARAM", "1"),
			AgentHost:    getEnv("JAEGER_AGENT_HOST", "localhost"),
			AgentPort:    getEnv("JAEGER_AGENT_PORT", "6831"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
