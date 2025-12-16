package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// S3Service handles AWS S3 operations for receipt files
type S3Service struct {
	client     *s3.Client
	bucketName string
}

// NewS3Service creates a new S3 service
func NewS3Service(region, accessKeyID, secretKey, bucketName string) (*S3Service, error) {
	// Create AWS config with static credentials
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &S3Service{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// GenerateFileKey creates an S3 object key (path) for a receipt file
// Format: {user_id}/{receipt_id}/{original_filename}
func (s *S3Service) GenerateFileKey(userID, receiptID, filename string) string {
	// Sanitize filename to avoid path traversal issues
	safeFilename := filepath.Base(filename)
	return fmt.Sprintf("%s/%s/%s", userID, receiptID, safeFilename)
}

// UploadFile uploads a file to S3 and returns the S3 key
func (s *S3Service) UploadFile(ctx context.Context, file io.Reader, fileSize int64, key, contentType string) error {
	// Create PutObject input
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileSize),
	}

	// Upload file
	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		// Check for specific AWS errors
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NoSuchBucket":
				return fmt.Errorf("S3 bucket '%s' does not exist. Please create the bucket first", s.bucketName)
			case "AccessDenied":
				return fmt.Errorf("access denied to S3 bucket '%s'. Check your AWS credentials and bucket permissions", s.bucketName)
			case "InvalidAccessKeyId":
				return fmt.Errorf("invalid AWS access key ID. Check your AWS_ACCESS_KEY_ID")
			case "SignatureDoesNotMatch":
				return fmt.Errorf("invalid AWS secret access key. Check your AWS_SECRET_ACCESS_KEY")
			default:
				return fmt.Errorf("S3 upload failed (error code: %s): %w", apiErr.ErrorCode(), err)
			}
		}
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for accessing a file
// The URL expires after the specified duration
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		// Check if it's a "not found" error - that's okay for idempotency
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return nil // File doesn't exist, but that's fine
			}
		}
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// FileExists checks if a file exists in S3
func (s *S3Service) FileExists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
