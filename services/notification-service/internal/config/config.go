package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
// Loaded from environment variables
type Config struct {
	// AWS configuration
	AWSRegion      string
	AWSAccessKeyID string
	AWSSecretKey   string

	// SQS Queue URLs (for consuming events)
	ExpenseEventsQueueURL string
	ReceiptEventsQueueURL string
	AuthEventsQueueURL    string

	// SNS Topic ARN (for sending email notifications)
	NotificationEmailTopicARN string

	// Server configuration
	ServerPort string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// AWS configuration
	cfg.AWSRegion = getEnv("AWS_REGION", "us-east-1")
	cfg.AWSAccessKeyID = getEnv("AWS_ACCESS_KEY_ID", "")
	cfg.AWSSecretKey = getEnv("AWS_SECRET_ACCESS_KEY", "")

	// Validate AWS configuration
	if cfg.AWSAccessKeyID == "" || cfg.AWSSecretKey == "" {
		return nil, fmt.Errorf("AWS credentials are required. Please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
	}

	// SQS Queue URLs
	cfg.ExpenseEventsQueueURL = getEnv("EXPENSE_EVENTS_QUEUE_URL", "")
	cfg.ReceiptEventsQueueURL = getEnv("RECEIPT_EVENTS_QUEUE_URL", "")
	cfg.AuthEventsQueueURL = getEnv("AUTH_EVENTS_QUEUE_URL", "")

	// SNS Email Topic ARN
	cfg.NotificationEmailTopicARN = getEnv("NOTIFICATION_EMAIL_TOPIC_ARN", "")
	if cfg.NotificationEmailTopicARN == "" {
		return nil, fmt.Errorf("NOTIFICATION_EMAIL_TOPIC_ARN environment variable is required")
	}

	// Server port (default: 8083 to avoid conflict with other services)
	cfg.ServerPort = getEnv("SERVER_PORT", "8083")

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
