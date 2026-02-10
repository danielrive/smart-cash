resource "aws_secretsmanager_secret" "secret" {
  name        = var.secret_name
  description = var.description
  
  tags = merge(
    {
      Environment = var.environment
      ManagedBy    = "Terraform"
      Project      = var.project_name
    },
    var.additional_tags
  )
}
