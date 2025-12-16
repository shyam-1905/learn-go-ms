package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"expense-tracker/notification-service/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSConsumer handles consuming messages from SQS queues
type SQSConsumer struct {
	client *sqs.Client
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer(region, accessKeyID, secretKey string) (*SQSConsumer, error) {
	// Create AWS config with static credentials
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SQS client
	client := sqs.NewFromConfig(cfg)

	return &SQSConsumer{
		client: client,
	}, nil
}

// MessageHandler is a function type for handling messages
type MessageHandler func(ctx context.Context, event *model.Event) error

// ConsumeMessages polls an SQS queue and processes messages
// It runs continuously until the context is cancelled
func (c *SQSConsumer) ConsumeMessages(ctx context.Context, queueURL string, handler MessageHandler) error {
	log.Printf("Starting to consume messages from queue: %s", queueURL)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			log.Println("Stopping message consumption")
			return ctx.Err()
		default:
		}

		// Receive messages from queue
		// Max 10 messages per batch (SQS limit)
		result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // Long polling (wait up to 20 seconds for messages)
			VisibilityTimeout:   30, // Hide message for 30 seconds while processing
		})

		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// Process each message
		for _, message := range result.Messages {
			if err := c.processMessage(ctx, message, queueURL, handler); err != nil {
				log.Printf("Error processing message: %v", err)
				// Don't delete message on error - let it become visible again for retry
				continue
			}

			// Delete message after successful processing
			_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
		}

		// If no messages, the loop continues (long polling will wait)
	}
}

// processMessage processes a single SQS message
func (c *SQSConsumer) processMessage(ctx context.Context, message types.Message, queueURL string, handler MessageHandler) error {
	// SQS messages from SNS subscriptions have the message body wrapped in an SNS envelope
	// We need to extract the actual event data

	var snsEnvelope struct {
		Type             string `json:"Type"`
		MessageId        string `json:"MessageId"`
		TopicArn         string `json:"TopicArn"`
		Message          string `json:"Message"` // This contains our actual event JSON
		Timestamp        string `json:"Timestamp"`
		SignatureVersion string `json:"SignatureVersion"`
		Signature        string `json:"Signature"`
		SigningCertURL   string `json:"SigningCertURL"`
		UnsubscribeURL   string `json:"UnsubscribeURL"`
	}

	// Parse SNS envelope
	if err := json.Unmarshal([]byte(*message.Body), &snsEnvelope); err != nil {
		return fmt.Errorf("failed to parse SNS envelope: %w", err)
	}

	// Check if this is an SNS notification
	if snsEnvelope.Type == "Notification" {
		// Parse the actual event from the Message field
		var event model.Event
		if err := json.Unmarshal([]byte(snsEnvelope.Message), &event); err != nil {
			return fmt.Errorf("failed to parse event message: %w", err)
		}

		// Call the handler
		if err := handler(ctx, &event); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}

		log.Printf("Successfully processed event: %s for user %s", event.EventType, event.UserID)
	} else {
		// If not an SNS notification, try to parse directly as event
		var event model.Event
		if err := json.Unmarshal([]byte(*message.Body), &event); err != nil {
			return fmt.Errorf("failed to parse event: %w", err)
		}

		// Call the handler
		if err := handler(ctx, &event); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}

		log.Printf("Successfully processed event: %s for user %s", event.EventType, event.UserID)
	}

	return nil
}
