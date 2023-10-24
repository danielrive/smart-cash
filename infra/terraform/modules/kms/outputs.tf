output "kms_arn" {
  description = "The ID of the secret manager"
  value       = try(aws_kms_key.key.arn, "")
}

output "kms_id" {
  description = "The ID of the secret manager"
  value       = try(aws_kms_key.key.key_id, "")
}

output "kms_alias" {
  description = "Alias of the KMS key"
  value       = aws_kms_alias.key_alias.arn
}