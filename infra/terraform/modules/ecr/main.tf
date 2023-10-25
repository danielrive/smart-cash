### ECR Creation

resource "aws_ecr_repository" "this" {
  name                 = "${var.name}-${var.project_name}"
  image_tag_mutability = "MUTABLE"
  force_delete         = var.force_delete
  image_scanning_configuration {
    scan_on_push = true
  }
  encryption_configuration {
    encryption_type = "AES256"
  }
}