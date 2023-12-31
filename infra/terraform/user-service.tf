################################################
########## Resources for User-service

#######################
#### DynamoDB tables

### Users Table

resource "aws_dynamodb_table" "user_table" {
  name         = "user-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "userId"
  range_key    = "email"

  attribute {
    name = "userId"
    type = "N"
  }

  attribute {
    name = "email"
    type = "S"
  }

  attribute {
    name = "username"
    type = "S"
  }

  global_secondary_index {
    name               = "by_email"
    hash_key           = "email"
    projection_type    = "INCLUDE"
    non_key_attributes = ["userId","email","status","username"]
  }

  global_secondary_index {
    name               = "by_username"
    hash_key           = "username"
    projection_type    = "INCLUDE"
    non_key_attributes = ["userId","email","status","username"]
  }
  tags = {
    Name = "users_${var.environment}"
  }
}


########################
###### IAM Role

resource "aws_iam_role" "user-service-role" {
  name = "role-user-service-${var.environment}"
  path = "/"
  assume_role_policy = jsonencode({
  Version="2012-10-17"
  Statement =  [
    {
      Effect= "Allow"
      Principal= {
        Federated= "arn:aws:iam::${data.aws_caller_identity.id_account.id}:oidc-provider/${module.eks_cluster.cluster_oidc}"
      },
      Action= "sts:AssumeRoleWithWebIdentity",
      Condition={
        StringEquals= {
          "${module.eks_cluster.cluster_oidc}:aud": "sts.amazonaws.com",
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-user-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA user-service

resource "aws_iam_policy" "dynamodb-user-service-policy" {
  name        = "policy-dynamodb-user-service-${var.environment}"
  path        = "/"
  description = "policy for k8 service account"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
                "dynamodb:ConditionCheckItem",
                "dynamodb:PutItem",
                "dynamodb:DescribeTable",
                "dynamodb:DeleteItem",
                "dynamodb:GetItem",
                "dynamodb:Query",
                "dynamodb:UpdateItem"
        ]
        Effect   = "Allow"
        Resource = aws_dynamodb_table.user_table.arn
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-user-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-user-service-policy.arn
  role       = aws_iam_role.user-service-role.name
}

#######################
####  ECR Repo

module "ecr_registry_user_service" {
  source       = "./modules/ecr"
  name         = "user-service"
  project_name = var.project_name
  environment  = var.environment
}