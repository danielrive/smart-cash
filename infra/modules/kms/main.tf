resource "aws_kms_key" "key" {
  description             = "KMS key  for ${var.name} ${var.project_name}-${var.environment} in ${var.region} "
  deletion_window_in_days = var.deletion_days
  enable_key_rotation     = var.enable_key_rotation
  policy                  = var.key_policy
  tags = {
    Name = "kms-${var.name}-${var.project_name}-${var.environment}-${var.region}"
  }
}

resource "aws_kms_alias" "key_alias" {
  name          = "alias/kms-${var.name}-${var.project_name}-${var.environment}-${var.region}"
  target_key_id = aws_kms_key.key.key_id
}