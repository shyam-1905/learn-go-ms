# Notification Service

The notification service handles email notifications for the expense tracker microservices architecture. It consumes events from SQS queues and sends email notifications via AWS SNS.

## Architecture

- **Event-Driven**: Consumes events from SQS queues (expense, receipt, auth events)
- **Email Notifications**: Sends emails via AWS SNS
- **Template-Based**: Uses HTML email templates for different event types

## Setup

### 1. AWS Resources Setup

Run the setup script to create SNS topics and SQS queues:

```powershell
.\set-env.ps1
go run ./cmd/setup-aws.go
```

This will create:
- **SNS Topics**: `expense-events-topic`, `receipt-events-topic`, `auth-events-topic`, `notification-email-topic`
- **SQS Queues**: `expense-events-queue`, `receipt-events-queue`, `auth-events-queue`
- **Subscriptions**: Queues subscribed to their respective topics

### 2. Environment Variables

Update `set-env.ps1` with your AWS credentials and the queue URLs/topic ARNs from the setup script output.

### 3. Email Subscription

Users need to subscribe their email addresses to the `notification-email-topic` to receive notifications. This can be done via:
- AWS SNS Console
- API call: `SubscribeEmail` method in the service

### 4. Run the Service

```powershell
.\set-env.ps1
go run ./cmd/main.go
```

## Event Types

The service handles the following events:

- `expense.created` - When an expense is created
- `expense.updated` - When an expense is updated
- `receipt.uploaded` - When a receipt is uploaded
- `receipt.linked` - When a receipt is linked to an expense
- `user.registered` - When a new user registers

## Endpoints

- `GET /health` - Health check endpoint

## Configuration

Required environment variables:

- `AWS_REGION` - AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `EXPENSE_EVENTS_QUEUE_URL` - SQS queue URL for expense events
- `RECEIPT_EVENTS_QUEUE_URL` - SQS queue URL for receipt events
- `AUTH_EVENTS_QUEUE_URL` - SQS queue URL for auth events
- `NOTIFICATION_EMAIL_TOPIC_ARN` - SNS topic ARN for sending emails
- `SERVER_PORT` - Server port (default: 8083)

## Integration

Other services (auth-service, expense-service, receipt-service) publish events to SNS topics, which are automatically forwarded to SQS queues and consumed by this service.

Make sure to configure the topic ARNs in each service's `set-env.ps1`:
- `expense-service`: `EXPENSE_EVENTS_TOPIC_ARN`
- `receipt-service`: `RECEIPT_EVENTS_TOPIC_ARN`
- `auth-service`: `AUTH_EVENTS_TOPIC_ARN`
