# Email Notification Guide

## How Email Notifications Work

### Current Implementation

The notification service sends emails to the address specified in `event.UserEmail` from the events it receives. Here's how it works:

1. **Event Publishing**: Services (auth, expense, receipt) publish events with `UserEmail` field
2. **Event Consumption**: Notification service consumes events from SQS queues
3. **Email Sending**: Notification service extracts `event.UserEmail` and sends email via AWS SNS

### Email Address Source

- **`user.registered` events**: Email is included (from user registration)
- **`expense.created/updated` events**: Email is extracted from JWT token context (auth middleware)
- **`receipt.uploaded/linked` events**: Email is extracted from JWT token context (auth middleware)

### Important: AWS SNS Email Subscription Requirement

**⚠️ Critical**: AWS SNS requires email addresses to be **subscribed and confirmed** before they can receive emails.

#### How SNS Email Subscriptions Work:

1. **Subscribe**: Email address must be subscribed to the SNS topic
2. **Confirm**: User receives a confirmation email and must click the confirmation link
3. **Receive**: Only after confirmation, the email can receive notifications

### Steps to Enable Email Notifications

#### Option 1: Subscribe via AWS Console

1. Go to AWS SNS Console
2. Select the topic: `notification-email-topic`
3. Click "Create subscription"
4. Choose protocol: "Email"
5. Enter email address
6. Click "Create subscription"
7. Check email inbox and confirm subscription

#### Option 2: Subscribe via API

Use the `SubscribeEmail` method in the notification service:

```go
// Example: Subscribe an email
subscriptionARN, err := snsPublisher.SubscribeEmail(ctx, "user@example.com")
```

Then the user must confirm via the confirmation email sent by AWS SNS.

### Current Email Flow

```
Event Published → SNS Topic → SQS Queue → Notification Service
                                                      ↓
                                              Extract UserEmail
                                                      ↓
                                              Send to SNS Topic
                                                      ↓
                                              (Only if subscribed)
                                                      ↓
                                              User's Email Inbox
```

### Testing Email Notifications

1. **Subscribe your email** to `notification-email-topic` (via console or API)
2. **Confirm the subscription** (click link in confirmation email)
3. **Trigger an event**:
   - Register a new user → `user.registered` event
   - Create an expense → `expense.created` event
   - Upload a receipt → `receipt.uploaded` event
4. **Check your email** for the notification

### Troubleshooting

**No emails received?**
- Check if email is subscribed to the topic
- Check if subscription is confirmed (not pending)
- Check notification service logs for errors
- Verify `event.UserEmail` is not empty in logs

**Email is empty in events?**
- Ensure auth middleware is extracting email from JWT token
- Check that services are passing context with user_email

### Future Improvements

Consider implementing:
1. **Automatic subscription**: Subscribe user email when they register
2. **Email validation**: Verify email format before sending
3. **Fallback handling**: Handle cases where email is not available
4. **Direct email sending**: Use AWS SES instead of SNS for more control
