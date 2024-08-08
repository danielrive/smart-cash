################################################
########## Resources for expenses-service

locals {
  this_service_name     = "expenses"
  this_service_port     = 8282
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo     = var.environment
}

#######################
#### DynamoDB tables

### expenses Table

resource "aws_dynamodb_table" "dynamo_table" {
  name         = "${local.this_service_name}-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "expenseId"

  attribute {
    name = "expenseId"
    type = "S"
  }

  attribute {
    name = "category"
    type = "S"
  }

  attribute {

    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name               = "by_userId"
    hash_key           = "userId"
    projection_type    = "INCLUDE"
    non_key_attributes = ["expenseId", "currency", "date", "amount"]
  }

  global_secondary_index {
    name            = "by_category"
    hash_key        = "userId"
    range_key       = "category"
    projection_type = "ALL"
  }
  tags = {
    Name = "${local.this_service_name}-table"
  }

}


##############################
###### IAM Role K8 SA

resource "aws_iam_role" "iam_sa_role" {
  name = "role-${local.this_service_name}-${var.environment}"
  path = "/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.id_account.id}:oidc-provider/${data.terraform_remote_state.eks.outputs.cluster_oidc}"
        },
        Action = "sts:AssumeRoleWithWebIdentity",
        Condition = {
          StringEquals = {
            "${data.terraform_remote_state.eks.outputs.cluster_oidc}:aud" : "sts.amazonaws.com",
            "${data.terraform_remote_state.eks.outputs.cluster_oidc}:sub" : "system:serviceaccount:${var.environment}:sa-${local.this_service_name}-service"
          }
        }
      }
    ]
  })
}

####### IAM policy for SA expenses

resource "aws_iam_policy" "dynamodb_iam_policy" {
  name        = "policy-dynamodb-${local.this_service_name}-${var.environment}"
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
        Effect = "Allow"
        Resource = [
          aws_dynamodb_table.dynamo_table.arn,
          "${aws_dynamodb_table.dynamo_table.arn}/index/by_userId",
          "${aws_dynamodb_table.dynamo_table.arn}/index/by_category"
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "att_policy_role1" {
  policy_arn = aws_iam_policy.dynamodb_iam_policy.arn
  role       = aws_iam_role.iam_sa_role.name
}



#############################
##### ECR Repo

module "ecr_registry" {
  source       = "../../modules/ecr"
  name         = "${local.this_service_name}-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base_manifests" {
  for_each   = fileset("../microservices-templates", "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/base/${each.key}"
  content = templatefile(
    "../microservices-templates/${each.key}",
    {
      SERVICE_NAME               = local.this_service_name
      SERVICE_PORT               = local.this_service_port
      SERVICE_PATH_HEALTH_CHECKS = "health"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



###########################
##### overlays

resource "github_repository_file" "overlays_svc" {
  for_each   = fileset("${local.path_tf_repo_services}/overlays/${var.environment}", "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME        = local.this_service_name
      ECR_REPO            = module.ecr_registry.repo_url
      ARN_ROLE_SERVICE    = aws_iam_role.iam_sa_role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.dynamo_table.name
      AWS_REGION          = var.region
      ENVIRONMENT         = var.environment
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
##### Network Policies

resource "github_repository_file" "network_policy" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/${local.this_service_name}.yaml",
    {
      PROJECT_NAME = var.project_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

###########################
##### Images Updates automation

resource "github_repository_file" "image_updates" {
  for_each   = fileset("${local.path_tf_repo_services}/flux-image-update", "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/base/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/flux-image-update/${each.key}",
    {
      SERVICE_NAME    = local.this_service_name
      ECR_REPO        = module.ecr_registry.repo_url
      ENVIRONMENT     = var.environment
      PATH_DEPLOYMENT = "services/${local.this_service_name}-service/overlays/${var.environment}/kustomization.yaml"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}