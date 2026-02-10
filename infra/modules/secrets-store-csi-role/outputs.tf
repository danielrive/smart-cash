output "role_arn" {
  description = "ARN of the IAM role for Secrets Store CSI Driver AWS Provider"
  value       = aws_iam_role.pod_sa_role.arn
}
