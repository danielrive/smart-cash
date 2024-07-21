################################################
########## Resources for User-service
locals {
  this_service_name = "user"
  this_service_port = 8181
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo = var.environment
}


#######################
#### DynamoDB tables

### Users Table

resource "aws_dynamodb_table" "user_table" {
  name         = "user-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "userId"

  attribute {
    name = "userId"
    type = "S"
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
    non_key_attributes = ["userId","email","status","username","password"]
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

resource "aws_iam_role" "user-role" {
  name = "role-user-${var.environment}"
  path = "/"
  assume_role_policy = jsonencode({
  Version="2012-10-17"
  Statement =  [
    {
      Effect= "Allow"
      Principal= {
        Federated= "arn:aws:iam::${data.aws_caller_identity.id_account.id}:oidc-provider/${data.terraform_remote_state.eks.outputs.cluster_oidc}"
      },
      Action= "sts:AssumeRoleWithWebIdentity",
      Condition={
        StringEquals= {
          "${data.terraform_remote_state.eks.outputs.cluster_oidc}:aud": "sts.amazonaws.com",
          "${data.terraform_remote_state.eks.outputs.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-user-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA user

resource "aws_iam_policy" "dynamodb-user-policy" {
  name        = "policy-dynamodb-user-${var.environment}"
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
        Resource = [
                    aws_dynamodb_table.user_table.arn,
                    "${aws_dynamodb_table.user_table.arn}/index/by_email",
                    "${aws_dynamodb_table.user_table.arn}/index/by_username"
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-user-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-user-policy.arn
  role       = aws_iam_role.user-role.name
}

#######################
####  ECR Repo

module "ecr_registry_user_service" {
  source       = "../../modules/ecr"
  name         = "user-service"
  project_name = var.project_name
  environment  = var.environment
}

###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests" {
  for_each            = fileset("../microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "services/user-service/base/${each.key}"
  content = templatefile(
    "../microservices-templates/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      SERVICE_PORT = local.this_service_port
      SERVICE_PATH_HEALTH_CHECKS = "/health"     
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



###########################
##### overlays

resource "github_repository_file" "overlays-user-svc" {
  for_each            = fileset("${local.path_tf_repo_services}/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "services/user-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      ECR_REPO = module.ecr_registry_user_service.repo_url
      ARN_ROLE_SERVICE = aws_iam_role.user-role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.user_table.name
      AWS_REGION  = var.region
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

###########################
##### Network Policies
/*
resource "github_repository_file" "np-user" {
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "services/user-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/user.yaml",
    {
      PROJECT_NAME  = var.project_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
*/