output "role_name" {
  value = aws_iam_role.pod_sa_role.name
}

output "role_arn" {
  value = aws_iam_role.pod_sa_role.arn
}