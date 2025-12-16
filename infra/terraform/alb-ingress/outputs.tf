output "alb_controller_role_arn" {
  description = "ARN of IAM role for AWS Load Balancer Controller"
  value       = aws_iam_role.alb_controller.arn
}

output "alb_controller_service_account_name" {
  description = "Name of the Kubernetes service account for ALB controller"
  value       = kubernetes_service_account.alb_controller.metadata[0].name
}

