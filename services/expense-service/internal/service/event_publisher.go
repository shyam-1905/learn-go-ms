package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// EventPublisher publishes events to SNS topics
type EventPublisher struct {
	client   *sns.Client
	topicARN string
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(region, accessKeyID, secretKey, topicARN string) (*EventPublisher, error) {
	// Create AWS config with static credentials
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SNS client
	client := sns.NewFromConfig(cfg)

	return &EventPublisher{
		client:   client,
		topicARN: topicARN,
	}, nil
}

// Event represents an event to be published
type Event struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	UserEmail string                 `json:"user_email"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishEvent publishes an event to the SNS topic
// This is non-blocking - errors are logged but don't affect the main operation
func (p *EventPublisher) PublishEvent(ctx context.Context, event *Event) error {
	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to SNS topic
	input := &sns.PublishInput{
		TopicArn: aws.String(p.topicARN),
		Message:  aws.String(string(eventJSON)),
		Subject:  aws.String(event.EventType),
	}

	result, err := p.client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Successfully published event %s to topic %s (MessageId: %s)", event.EventType, p.topicARN, *result.MessageId)
	return nil
}

// PublishEventAsync publishes an event asynchronously (non-blocking)
// Uses a background context to avoid cancellation when the HTTP request completes
func (p *EventPublisher) PublishEventAsync(ctx context.Context, event *Event) {
	go func() {
		// Use background context to avoid cancellation when HTTP request completes
		bgCtx := context.Background()
		log.Printf("Publishing event %s for user %s (email: %s)", event.EventType, event.UserID, event.UserEmail)
		if err := p.PublishEvent(bgCtx, event); err != nil {
			// Log error but don't fail the main operation
			log.Printf("ERROR: Failed to publish event %s: %v", event.EventType, err)
		}
	}()
}
