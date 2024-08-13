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
  policy = <<EOF
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

resource "aws_ecr_registry_policy" "allow_pod_pull" {
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "alloPull",
        Effect = "Allow",
        Principal = {
          "AWS" : "${var.service_role}"
        },
        Action = [
                "ecr:BatchCheckLayerAvailability",
                "ecr:BatchGetImage",
                "ecr:GetDownloadUrlForLayer"
              ],
        Resource = [
          aws_ecr_repository.this.arn
        ]
      },
      {
        Sid    = "alloPush",
        Effect = "Allow",
        Principal = {
          "AWS" : "arn:aws:iam::${var.account_id}:role/GitHubAction-smart-cash"
        },
        Action = [
                "ecr:BatchCheckLayerAvailability",
                "ecr:CompleteLayerUpload",
                "ecr:InitiateLayerUpload",
                "ecr:PutImage",
                "ecr:UploadLayerPart"
              ],
        Resource = [
          aws_ecr_repository.this.arn
        ]
      }

    ]
  })
}
