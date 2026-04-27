# Grants the workload's Pod Identity IAM role permission to read the shared JWT secret
# from Secrets Manager when the Secrets Store CSI driver mounts it (usePodIdentity: true).
data "aws_secretsmanager_secret" "jwt" {
  name = "smart-cash/jwt-secret/${var.environment}"
}

resource "aws_iam_role_policy" "jwt_secret_read" {
  name = "jwt-secret-read-${var.name_suffix}-${var.environment}"
  role = var.iam_role_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ReadJwtSecretForCsiMount"
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
        ]
        Resource = data.aws_secretsmanager_secret.jwt.arn
      }
    ]
  })
}
