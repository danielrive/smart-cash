output "role_arn" {
  description = "ARN of the IAM role for External Secrets Operator"
  value       = aws_iam_role.pod_sa_role.arn
}
