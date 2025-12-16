package main

import (
	"context"
	"errors"
	"expense-tracker/receipt-service/internal/config"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func createS3Bucket() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Create S3 Bucket")
	fmt.Println("========================================")
	fmt.Printf("AWS Region: %s\n", cfg.AWSRegion)
	fmt.Printf("Bucket Name: %s\n", cfg.S3BucketName)
	fmt.Println("")

	// Create AWS config with static credentials
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

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg)
	ctx := context.Background()

	// Check if bucket already exists
	fmt.Printf("Checking if bucket '%s' exists...\n", cfg.S3BucketName)
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3BucketName),
	})

	if err == nil {
		fmt.Printf("✅ Bucket '%s' already exists!\n", cfg.S3BucketName)
		fmt.Println("")
		fmt.Println("========================================")
		fmt.Println("Bucket Check: PASSED ✅")
		fmt.Println("========================================")
		return
	}

	// Bucket doesn't exist (or HeadBucket failed), try to create it
	fmt.Printf("Bucket '%s' does not exist or not accessible. Creating it...\n", cfg.S3BucketName)

	// Create bucket input
	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(cfg.S3BucketName),
	}

	// For regions other than us-east-1, we need to specify the location constraint
	if cfg.AWSRegion != "us-east-1" {
		// Convert region string to location constraint type
		// Note: For us-east-1, we don't set this at all
		locationConstraint := types.BucketLocationConstraint(cfg.AWSRegion)
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: locationConstraint,
		}
	}

	// Create the bucket
	_, err = s3Client.CreateBucket(ctx, createInput)
	if err != nil {
		// Check if bucket already exists (this can happen if HeadBucket didn't detect it)
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "BucketAlreadyExists" || apiErr.ErrorCode() == "BucketAlreadyOwnedByYou" {
				fmt.Printf("✅ Bucket '%s' already exists (detected during creation attempt)!\n", cfg.S3BucketName)
				fmt.Println("")
				fmt.Println("========================================")
				fmt.Println("Bucket Check: PASSED ✅")
				fmt.Println("========================================")
				return
			}
		}
		log.Fatalf("❌ Failed to create bucket: %v", err)
	}

	fmt.Printf("✅ Bucket '%s' created successfully!\n", cfg.S3BucketName)
	fmt.Println("")

	// Wait a moment for the bucket to be fully available
	fmt.Println("Waiting for bucket to be available...")

	// Set up bucket versioning (optional but recommended)
	fmt.Println("Configuring bucket settings...")

	// Enable versioning (optional - helps with data protection)
	_, err = s3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(cfg.S3BucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		// Versioning is optional, just log a warning
		fmt.Printf("⚠️  Warning: Could not enable versioning (this is optional): %v\n", err)
	} else {
		fmt.Println("✅ Versioning enabled")
	}

	// Block public access (security best practice)
	fmt.Println("Configuring public access settings...")
	_, err = s3Client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(cfg.S3BucketName),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not configure public access block: %v\n", err)
	} else {
		fmt.Println("✅ Public access blocked (security best practice)")
	}

	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("Bucket Created Successfully! ✅")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Printf("Bucket Name: %s\n", cfg.S3BucketName)
	fmt.Printf("Region: %s\n", cfg.AWSRegion)
	fmt.Println("")
	fmt.Println("You can now upload receipts to this bucket.")
	fmt.Println("Run the service with: go run cmd/main.go")
}
