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

// Sleep, this was necessary because the role for the service account take some time to be able to use in the console
// the role is created by terraform but for some reason the ECR policy doesnt see yet 

### Force to update the Pod to take the changes in the SA
resource "null_resource" "force_to_wait" {
  depends_on = [null_resource.bootstrap-flux,github_repository_file.patch_flux]
  provisioner "local-exec" {
    command = <<EOF
    sleep 2
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }
}

//  IAM Policy for repository, just allow pull for specific roles
resource "aws_ecr_repository_policy" "allow_pod_pull" {
  depends_on = [aws_ecr_repository.this,null_resource.restart-image-reflector]
  repository = aws_ecr_repository.this.name
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AllowPull",
        Effect = "Allow",
        Principal = {
          "AWS" : [
            "${var.service_role}",
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
