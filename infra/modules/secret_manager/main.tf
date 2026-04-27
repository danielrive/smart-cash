resource "aws_secretsmanager_secret" "secret" {
  name        = var.secret_name
  description = var.description

  tags = merge(
    {
      Environment = var.environment
      ManagedBy   = "Terraform"
      Project     = var.project_name
    },
    var.additional_tags
  )
}

# Bootstrap so GetSecretValue works before a human sets the real JWT (rotate in AWS console; TF ignores changes after).
resource "random_password" "jwt_bootstrap" {
  count   = var.create_bootstrap_secret_version ? 1 : 0
  length  = 48
  special = true
}

resource "aws_secretsmanager_secret_version" "bootstrap" {
  count         = var.create_bootstrap_secret_version ? 1 : 0
  secret_id     = aws_secretsmanager_secret.secret.id
  secret_string = random_password.jwt_bootstrap[0].result

  lifecycle {
    ignore_changes = [secret_string]
  }
}
