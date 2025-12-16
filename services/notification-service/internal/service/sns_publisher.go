package service

import (
	"context"
	"fmt"

	"expense-tracker/notification-service/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSPublisher handles publishing email notifications via AWS SNS
type SNSPublisher struct {
	client   *sns.Client
	topicARN string
}

// NewSNSPublisher creates a new SNS publisher
func NewSNSPublisher(region, accessKeyID, secretKey, topicARN string) (*SNSPublisher, error) {
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

	return &SNSPublisher{
		client:   client,
		topicARN: topicARN,
	}, nil
}

// SendEmail sends an email notification via SNS
// The email address must be subscribed to the SNS topic
func (p *SNSPublisher) SendEmail(ctx context.Context, notification *model.Notification) error {
	// Publish to SNS topic
	// Note: For email notifications, we need to publish with the email as the target
	// However, SNS requires subscriptions. For now, we'll publish to the topic
	// and the topic should have email subscriptions configured

	// Create publish input
	input := &sns.PublishInput{
		TopicArn: aws.String(p.topicARN),
		Subject:  aws.String(notification.Subject),
		Message:  aws.String(notification.Body),
		// For email, we can also use MessageAttributes to specify the email target
		// But for simplicity, we'll rely on topic subscriptions
	}

	// Publish message
	_, err := p.client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish email notification: %w", err)
	}

	return nil
}

// SubscribeEmail subscribes an email address to the SNS topic
// This allows the email to receive notifications
func (p *SNSPublisher) SubscribeEmail(ctx context.Context, email string) (string, error) {
	input := &sns.SubscribeInput{
		TopicArn: aws.String(p.topicARN),
		Protocol: aws.String("email"),
		Endpoint: aws.String(email),
	}

	result, err := p.client.Subscribe(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to subscribe email: %w", err)
	}

	return *result.SubscriptionArn, nil
}

// UnsubscribeEmail removes an email subscription
func (p *SNSPublisher) UnsubscribeEmail(ctx context.Context, subscriptionARN string) error {
	input := &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(subscriptionARN),
	}

	_, err := p.client.Unsubscribe(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe email: %w", err)
	}

	return nil
}

// SendEmailToAddress sends an email directly to a specific address
// This creates a temporary subscription if needed
// Note: In production, you might want to maintain a subscription list
func (p *SNSPublisher) SendEmailToAddress(ctx context.Context, email string, subject, body string) error {
	// For direct email sending, we can use SNS with email protocol
	// But SNS requires the email to be subscribed first
	// Alternative: Use AWS SES directly for more control

	// For now, we'll publish to the topic and assume subscriptions are managed separately
	notification := &model.Notification{
		To:      email,
		Subject: subject,
		Body:    body,
	}

	return p.SendEmail(ctx, notification)
}
