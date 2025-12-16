output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.receipts.id
}

output "bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.receipts.arn
}

output "bucket_domain_name" {
  description = "Domain name of the S3 bucket"
  value       = aws_s3_bucket.receipts.bucket_domain_name
}

