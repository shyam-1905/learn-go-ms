package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
// Loaded from environment variables
type Config struct {
	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Auth Service configuration
	// AUTH_SERVICE_URL is the base URL of auth-service for token validation
	AuthServiceURL string

	// AWS SNS configuration for event publishing
	AWSRegion             string
	AWSAccessKeyID        string
	AWSSecretKey          string
	ExpenseEventsTopicARN string

	// Server configuration
	ServerPort string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Database configuration
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	cfg.DBPort = getEnvAsInt("DB_PORT", 5432)
	cfg.DBUser = getEnv("DB_USER", "postgres")
	cfg.DBPassword = getEnv("DB_PASSWORD", "")
	cfg.DBName = getEnv("DB_NAME", "expense_db")

	// Auth Service configuration
	// This is the base URL of auth-service (e.g., "http://localhost:8080")
	cfg.AuthServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8080")
	if cfg.AuthServiceURL == "" {
		return nil, fmt.Errorf("AUTH_SERVICE_URL environment variable is required")
	}

	// AWS SNS configuration (optional - for event publishing)
	cfg.AWSRegion = getEnv("AWS_REGION", "us-east-1")
	cfg.AWSAccessKeyID = getEnv("AWS_ACCESS_KEY_ID", "")
	cfg.AWSSecretKey = getEnv("AWS_SECRET_ACCESS_KEY", "")
	cfg.ExpenseEventsTopicARN = getEnv("EXPENSE_EVENTS_TOPIC_ARN", "")
	// Note: AWS credentials and topic ARN are optional - events won't be published if not configured

	// Server port (default: 8081 to avoid conflict with auth-service on 8080)
	cfg.ServerPort = getEnv("SERVER_PORT", "8081")

	return cfg, nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetDatabaseURL constructs a PostgreSQL connection string
// Uses sslmode=require for AWS RDS
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
