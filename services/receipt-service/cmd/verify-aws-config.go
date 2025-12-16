package main

import (
	"context"
	"expense-tracker/receipt-service/internal/config"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func verifyAWSConfig() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("AWS Configuration Verification")
	fmt.Println("========================================")
	fmt.Printf("AWS Region: %s\n", cfg.AWSRegion)
	fmt.Printf("AWS Access Key ID: %s...\n", cfg.AWSAccessKeyID[:min(8, len(cfg.AWSAccessKeyID))])
	fmt.Printf("S3 Bucket Name: %s\n", cfg.S3BucketName)
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

	// Verify credentials by calling STS GetCallerIdentity
	fmt.Println("Verifying AWS credentials...")
	stsClient := sts.NewFromConfig(awsCfg)
	identity, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("❌ Failed to verify AWS credentials: %v", err)
	}
	fmt.Printf("✅ AWS credentials are valid!\n")
	fmt.Printf("   Account ID: %s\n", *identity.Account)
	fmt.Printf("   User ARN: %s\n", *identity.Arn)
	fmt.Println("")

	// List all buckets to see what we have access to
	fmt.Println("Listing all S3 buckets in your account...")
	s3Client := s3.NewFromConfig(awsCfg)
	buckets, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("❌ Failed to list buckets: %v", err)
	}

	fmt.Printf("Found %d bucket(s):\n", len(buckets.Buckets))
	if len(buckets.Buckets) == 0 {
		fmt.Println("  (no buckets found)")
	} else {
		for _, bucket := range buckets.Buckets {
			fmt.Printf("  - %s (created: %s)\n", *bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println("")

	// Check if our target bucket exists
	fmt.Printf("Checking if bucket '%s' exists...\n", cfg.S3BucketName)
	bucketExists := false
	for _, bucket := range buckets.Buckets {
		if *bucket.Name == cfg.S3BucketName {
			bucketExists = true
			fmt.Printf("✅ Bucket '%s' found in your account!\n", cfg.S3BucketName)

			// Get bucket location
			location, err := s3Client.GetBucketLocation(context.Background(), &s3.GetBucketLocationInput{
				Bucket: aws.String(cfg.S3BucketName),
			})
			if err == nil {
				region := "us-east-1" // Default region
				if location.LocationConstraint != "" {
					region = string(location.LocationConstraint)
				}
				fmt.Printf("   Region: %s\n", region)
				if region != cfg.AWSRegion {
					fmt.Printf("   ⚠️  WARNING: Bucket is in region '%s' but you're configured for '%s'!\n", region, cfg.AWSRegion)
					fmt.Printf("   Update AWS_REGION in set-env.ps1 to match the bucket region.\n")
				}
			}
			break
		}
	}

	if !bucketExists {
		fmt.Printf("❌ Bucket '%s' NOT found in your account.\n", cfg.S3BucketName)
		fmt.Println("")
		fmt.Println("The bucket needs to be created. Run:")
		fmt.Println("  go run cmd/create-s3-bucket.go")
	}

	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("Verification Complete")
	fmt.Println("========================================")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
