output "secret_arn" {
  description = "ARN of the secret in AWS Secrets Manager"
  value       = aws_secretsmanager_secret.secret.arn
}

output "secret_name" {
  description = "Name of the secret in AWS Secrets Manager"
  value       = aws_secretsmanager_secret.secret.name
}
