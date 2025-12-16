package main

import (
	"context"
	"expense-tracker/notification-service/internal/config"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func setupAWS() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("AWS SNS/SQS Setup for Notification Service")
	fmt.Println("========================================")
	fmt.Printf("AWS Region: %s\n", cfg.AWSRegion)
	fmt.Println("")

	// Create AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create clients
	snsClient := sns.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)
	ctx := context.Background()

	// SNS Topics to create
	topics := map[string]string{
		"expense-events-topic":     "Events from expense-service",
		"receipt-events-topic":     "Events from receipt-service",
		"auth-events-topic":        "Events from auth-service",
		"notification-email-topic": "Email notifications topic",
	}

	// SQS Queues to create (one per service topic)
	queues := []string{
		"expense-events-queue",
		"receipt-events-queue",
		"auth-events-queue",
	}

	topicARNs := make(map[string]string)
	queueURLs := make(map[string]string)

	// Create SNS topics
	fmt.Println("Creating SNS topics...")
	for topicName := range topics {
		fmt.Printf("Creating topic: %s...\n", topicName)

		// Check if topic already exists by trying to create it
		createInput := &sns.CreateTopicInput{
			Name: aws.String(topicName),
		}

		result, err := snsClient.CreateTopic(ctx, createInput)
		if err != nil {
			log.Fatalf("Failed to create topic %s: %v", topicName, err)
		}

		topicARNs[topicName] = *result.TopicArn
		fmt.Printf("✅ Topic '%s' created: %s\n", topicName, *result.TopicArn)
	}
	fmt.Println("")

	// Create SQS queues
	fmt.Println("Creating SQS queues...")
	for _, queueName := range queues {
		fmt.Printf("Creating queue: %s...\n", queueName)

		createInput := &sqs.CreateQueueInput{
			QueueName: aws.String(queueName),
		}

		result, err := sqsClient.CreateQueue(ctx, createInput)
		if err != nil {
			log.Fatalf("Failed to create queue %s: %v", queueName, err)
		}

		queueURLs[queueName] = *result.QueueUrl
		fmt.Printf("✅ Queue '%s' created: %s\n", queueName, *result.QueueUrl)
	}
	fmt.Println("")

	// Get queue ARNs (needed for SNS subscription)
	fmt.Println("Getting queue ARNs...")
	queueARNs := make(map[string]string)
	for queueName, queueURL := range queueURLs {
		attrs, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(queueURL),
			AttributeNames: []types.QueueAttributeName{"QueueArn"},
		})
		if err != nil {
			log.Fatalf("Failed to get queue ARN for %s: %v", queueName, err)
		}
		queueARNs[queueName] = attrs.Attributes["QueueArn"]
		fmt.Printf("✅ Queue ARN for '%s': %s\n", queueName, queueARNs[queueName])
	}
	fmt.Println("")

	// Subscribe SQS queues to SNS topics
	fmt.Println("Subscribing SQS queues to SNS topics...")

	// Map queues to topics
	queueTopicMap := map[string]string{
		"expense-events-queue": "expense-events-topic",
		"receipt-events-queue": "receipt-events-topic",
		"auth-events-queue":    "auth-events-topic",
	}

	for queueName, topicName := range queueTopicMap {
		fmt.Printf("Subscribing %s to %s...\n", queueName, topicName)

		topicARN := topicARNs[topicName]
		queueARN := queueARNs[queueName]

		// Subscribe queue to topic
		subscribeInput := &sns.SubscribeInput{
			TopicArn: aws.String(topicARN),
			Protocol: aws.String("sqs"),
			Endpoint: aws.String(queueARN),
		}

		_, err := snsClient.Subscribe(ctx, subscribeInput)
		if err != nil {
			log.Fatalf("Failed to subscribe %s to %s: %v", queueName, topicName, err)
		}

		// Set queue policy to allow SNS to send messages
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"Service": "sns.amazonaws.com"
					},
					"Action": "sqs:SendMessage",
					"Resource": "%s",
					"Condition": {
						"ArnEquals": {
							"aws:SourceArn": "%s"
						}
					}
				}
			]
		}`, queueARN, topicARN)

		_, err = sqsClient.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
			QueueUrl: aws.String(queueURLs[queueName]),
			Attributes: map[string]string{
				"Policy": policy,
			},
		})
		if err != nil {
			log.Fatalf("Failed to set queue policy for %s: %v", queueName, err)
		}

		fmt.Printf("✅ Subscribed %s to %s\n", queueName, topicName)
	}
	fmt.Println("")

	// Print summary
	fmt.Println("========================================")
	fmt.Println("Setup Complete! ✅")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Println("SNS Topics:")
	for topicName, topicARN := range topicARNs {
		fmt.Printf("  %s: %s\n", topicName, topicARN)
	}
	fmt.Println("")
	fmt.Println("SQS Queues:")
	for queueName, queueURL := range queueURLs {
		fmt.Printf("  %s: %s\n", queueName, queueURL)
	}
	fmt.Println("")
	fmt.Println("Environment Variables to set:")
	fmt.Printf("  EXPENSE_EVENTS_QUEUE_URL=%s\n", queueURLs["expense-events-queue"])
	fmt.Printf("  RECEIPT_EVENTS_QUEUE_URL=%s\n", queueURLs["receipt-events-queue"])
	fmt.Printf("  AUTH_EVENTS_QUEUE_URL=%s\n", queueURLs["auth-events-queue"])
	fmt.Printf("  NOTIFICATION_EMAIL_TOPIC_ARN=%s\n", topicARNs["notification-email-topic"])
	fmt.Println("")
	fmt.Println("Note: Users need to subscribe their email addresses to the notification-email-topic")
	fmt.Println("      to receive email notifications. This can be done via the SNS console or API.")
}
