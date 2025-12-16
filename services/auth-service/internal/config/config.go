package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
// We load this from environment variables
// Environment variables are better than hardcoding because:
// 1. Different values for dev/staging/production
// 2. Secrets aren't in code (security!)
// 3. Easy to change without recompiling
type Config struct {
	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// JWT configuration
	JWTSecret     string
	JWTExpiration time.Duration // How long tokens are valid

	// AWS SNS configuration for event publishing
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretKey       string
	AuthEventsTopicARN string

	// Server configuration
	ServerPort string
}

// Load reads configuration from environment variables
// os.Getenv() reads environment variables
// If a required variable is missing, we return an error
func Load() (*Config, error) {
	cfg := &Config{}

	// Database configuration
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	cfg.DBPort = getEnvAsInt("DB_PORT", 5432)
	cfg.DBUser = getEnv("DB_USER", "postgres")
	cfg.DBPassword = getEnv("DB_PASSWORD", "")
	cfg.DBName = getEnv("DB_NAME", "auth_db")

	// JWT configuration
	cfg.JWTSecret = getEnv("JWT_SECRET", "")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// JWT expiration (default: 24 hours)
	expirationHours := getEnvAsInt("JWT_EXPIRATION_HOURS", 24)
	cfg.JWTExpiration = time.Duration(expirationHours) * time.Hour

	// AWS SNS configuration (optional - for event publishing)
	cfg.AWSRegion = getEnv("AWS_REGION", "us-east-1")
	cfg.AWSAccessKeyID = getEnv("AWS_ACCESS_KEY_ID", "")
	cfg.AWSSecretKey = getEnv("AWS_SECRET_ACCESS_KEY", "")
	cfg.AuthEventsTopicARN = getEnv("AUTH_EVENTS_TOPIC_ARN", "")
	// Note: AWS credentials and topic ARN are optional - events won't be published if not configured

	// Server port (default: 8080)
	cfg.ServerPort = getEnv("SERVER_PORT", "8080")

	return cfg, nil
}

// getEnv reads an environment variable or returns a default value
// Helper function to make code cleaner
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer
// strconv.Atoi converts string to int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// If conversion fails, return default
		return defaultValue
	}
	return value
}

// GetDatabaseURL constructs a PostgreSQL connection string
// Format: postgres://user:password@host:port/dbname?sslmode=require
// Uses sslmode=require for AWS RDS (which requires SSL connections)
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
