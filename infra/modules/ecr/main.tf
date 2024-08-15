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

resource "aws_ecr_lifecycle_policy" "mandatory-policy" {
  repository = aws_ecr_repository.this.name
  policy     = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Expire images without tag",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 2
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}

//  IAM Policy for repository, just allow pull for specific roles
resource "aws_ecr_repository_policy" "allow_pod_pull" {
  depends_on = [aws_ecr_repository.this]
  repository = aws_ecr_repository.this.name
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AllowPull",
        Effect = "Allow",
        Principal = {
          "AWS" : [
           # "${var.service_role}",
            "arn:aws:iam::${var.account_id}:role/flux-images-${var.environment}-${var.region}",
          ]
        },
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
          "ecr:ListImages",
          "ecr:DescribePullThroughCacheRules",
          "ecr:DescribeImages",
        ]
      },
      {
        Sid    = "AllowPush",
        Effect = "Allow",
        Principal = {
          "AWS" : "arn:aws:iam::${var.account_id}:role/GitHubAction-smart-cash"

        },
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:CompleteLayerUpload",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage",
          "ecr:UploadLayerPart",
        ]
      }

    ]
  })
}
