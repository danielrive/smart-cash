output "role_name" {
    value = aws_iam_role.cert_manager.name
}

output "role_arn" {
  value = aws_iam_role.cert_manager.arn
}