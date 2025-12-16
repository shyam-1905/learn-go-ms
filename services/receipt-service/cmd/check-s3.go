package main

import (
	"context"
	"expense-tracker/receipt-service/internal/config"
	"expense-tracker/receipt-service/internal/service"
	"fmt"
	"log"
)

func checkS3() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("S3 Configuration Check")
	fmt.Println("========================================")
	fmt.Printf("AWS Region: %s\n", cfg.AWSRegion)
	fmt.Printf("AWS Access Key ID: %s...\n", cfg.AWSAccessKeyID[:min(8, len(cfg.AWSAccessKeyID))])
	fmt.Printf("S3 Bucket Name: %s\n", cfg.S3BucketName)
	fmt.Println("")

	// Initialize S3 service
	fmt.Println("Initializing S3 service...")
	s3Service, err := service.NewS3Service(
		cfg.AWSRegion,
		cfg.AWSAccessKeyID,
		cfg.AWSSecretKey,
		cfg.S3BucketName,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize S3 service: %v", err)
	}
	fmt.Println("✅ S3 service initialized")

	// Test bucket access by checking if bucket exists
	ctx := context.Background()
	fmt.Printf("\nChecking if bucket '%s' exists and is accessible...\n", cfg.S3BucketName)

	// Try to check if a test key exists (this will fail if bucket doesn't exist or we don't have access)
	testKey := "test-connection-check"
	exists, err := s3Service.FileExists(ctx, testKey)
	if err != nil {
		// Check for specific errors
		if contains(err.Error(), "NoSuchBucket") {
			fmt.Printf("❌ ERROR: Bucket '%s' does not exist!\n", cfg.S3BucketName)
			fmt.Println("\nTo fix this:")
			fmt.Println("1. Create the bucket in AWS S3 console")
			fmt.Println("2. Or update S3_BUCKET_NAME environment variable to an existing bucket")
			log.Fatal(err)
		} else if contains(err.Error(), "AccessDenied") {
			fmt.Printf("❌ ERROR: Access denied to bucket '%s'!\n", cfg.S3BucketName)
			fmt.Println("\nTo fix this:")
			fmt.Println("1. Check your AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)")
			fmt.Println("2. Ensure the IAM user/role has permissions to access the bucket")
			fmt.Println("3. Required permissions: s3:PutObject, s3:GetObject, s3:DeleteObject, s3:ListBucket")
			log.Fatal(err)
		} else if contains(err.Error(), "InvalidAccessKeyId") {
			fmt.Printf("❌ ERROR: Invalid AWS Access Key ID!\n")
			fmt.Println("\nTo fix this:")
			fmt.Println("1. Check your AWS_ACCESS_KEY_ID environment variable")
			fmt.Println("2. Ensure the access key is correct and active")
			log.Fatal(err)
		} else if contains(err.Error(), "SignatureDoesNotMatch") {
			fmt.Printf("❌ ERROR: Invalid AWS Secret Access Key!\n")
			fmt.Println("\nTo fix this:")
			fmt.Println("1. Check your AWS_SECRET_ACCESS_KEY environment variable")
			fmt.Println("2. Ensure the secret key matches the access key ID")
			log.Fatal(err)
		} else {
			fmt.Printf("❌ ERROR: %v\n", err)
			log.Fatal(err)
		}
	}

	// If we get here, bucket exists and we have access
	fmt.Printf("✅ Bucket '%s' exists and is accessible\n", cfg.S3BucketName)
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("S3 Configuration Check: PASSED ✅")
	fmt.Println("========================================")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
