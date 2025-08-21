package config

import (
	"os"
	"strconv"
	"time"
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

// AuthConfig holds auth-specific configuration
type AuthConfig struct {
	JWTSecret      string
	HTTPPort       string
	GRPCPort       string
	DatabaseURL    string
	RedisURL       string
	TokenExpiry    time.Duration
	RefreshExpiry  time.Duration
}

// LoadAuthConfig loads auth service specific configuration
func LoadAuthConfig() *AuthConfig {
	tokenExpiry, _ := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_EXPIRY", "15m"))
	refreshExpiry, _ := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_EXPIRY", "7d"))

	return &AuthConfig{
		JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		HTTPPort:      getEnv("HTTP_PORT", "8081"),
		GRPCPort:      getEnv("GRPC_PORT", "9090"),
		DatabaseURL:   getEnv("AUTH_DATABASE_URL", getEnv("DATABASE_URL", "")),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		TokenExpiry:   tokenExpiry,
		RefreshExpiry: refreshExpiry,
	}
}

// FeedConfig holds feed-specific configuration (for future use)
type FeedConfig struct {
	HTTPPort       string
	GRPCPort       string
	DatabaseURL    string
	RedisURL       string
}

// LoadFeedConfig loads feed service specific configuration
func LoadFeedConfig() *FeedConfig {
	return &FeedConfig{
		HTTPPort:    getEnv("HTTP_PORT", "8083"),
		GRPCPort:    getEnv("GRPC_PORT", "9091"),
		DatabaseURL: getEnv("FEED_DATABASE_URL", getEnv("DATABASE_URL", "")),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
