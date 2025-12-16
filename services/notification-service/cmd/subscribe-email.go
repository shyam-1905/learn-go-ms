package main

import (
	"context"
	"expense-tracker/notification-service/internal/config"
	"flag"
	"fmt"
	"log"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func subscribeEmail() {
	// Parse command line arguments
	email := flag.String("email", "", "Email address to subscribe (required)")
	flag.Parse()

	if *email == "" {
		fmt.Println("Error: Email address is required")
		fmt.Println("Usage: go run subscribe-email.go -email=your-email@example.com")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Subscribe Email to SNS Topic")
	fmt.Println("========================================")
	fmt.Printf("Email: %s\n", *email)
	fmt.Printf("Topic ARN: %s\n", cfg.NotificationEmailTopicARN)
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

	// Create SNS client
	snsClient := sns.NewFromConfig(awsCfg)
	ctx := context.Background()

	// Subscribe email to topic
	fmt.Printf("Subscribing %s to topic...\n", *email)
	protocol := "email"
	input := &sns.SubscribeInput{
		TopicArn: &cfg.NotificationEmailTopicARN,
		Protocol: &protocol,
		Endpoint: email,
	}

	result, err := snsClient.Subscribe(ctx, input)
	if err != nil {
		log.Fatalf("Failed to subscribe email: %v", err)
	}

	fmt.Println("✓ Subscription created successfully!")
	fmt.Printf("Subscription ARN: %s\n", *result.SubscriptionArn)
	fmt.Println("")
	fmt.Println("⚠️  IMPORTANT: Check your email inbox for a confirmation email from AWS SNS")
	fmt.Println("   You must click the confirmation link to activate the subscription.")
	fmt.Println("   Until confirmed, you will NOT receive notifications.")
	fmt.Println("")
	fmt.Println("After confirmation, you will receive notifications for:")
	fmt.Println("  - User registration")
	fmt.Println("  - Expense creation/updates")
	fmt.Println("  - Receipt uploads and links")
}
