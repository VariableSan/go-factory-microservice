package config

import (
	"os"
	"strconv"
)

// Config holds common configuration values
type Config struct {
	Port        string
	Environment string
	LogLevel    string
	RedisURL    string
	JaegerURL   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JaegerURL:   getEnv("JAEGER_ENDPOINT", "http://localhost:14268"),
	}
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string
	MaxConnections  int
	MaxIdleTime     int
	ConnMaxLifetime int
}

// LoadDatabaseConfig loads database configuration
func LoadDatabaseConfig() *DatabaseConfig {
	maxConn, _ := strconv.Atoi(getEnv("DB_MAX_CONNECTIONS", "10"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_TIME", "300"))
	maxLifetime, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "3600"))

	return &DatabaseConfig{
		URL:             getEnv("DATABASE_URL", ""),
		MaxConnections:  maxConn,
		MaxIdleTime:     maxIdle,
		ConnMaxLifetime: maxLifetime,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
