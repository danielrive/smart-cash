output "repo_arn" {
  value = aws_ecr_repository.this.arn
}

output "repo_id" {
  value = aws_ecr_repository.this.registry_id
}


output "repo_url" {
  value = aws_ecr_repository.this.repository_url
}