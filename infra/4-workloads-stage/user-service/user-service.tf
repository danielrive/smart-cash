locals {
  this_service_name     = "user"
  this_service_port     = 8181
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo     = var.environment
  cluster_name          = "${var.project_name}-${var.environment}"
  tier                  = "backend"
}


#######################
#### DynamoDB tables

### Users Table

resource "aws_dynamodb_table" "dynamo_table" {
  name         = "${local.this_service_name}-table"
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
    non_key_attributes = ["userId", "email", "status", "username", "password"]
  }

  global_secondary_index {
    name               = "by_username"
    hash_key           = "username"
    projection_type    = "INCLUDE"
    non_key_attributes = ["userId", "email", "status", "username"]
  }
  tags = {
    Name = "${local.this_service_name}-table"
  }
}

##############################
###### IAM Role K8 SA

resource "aws_iam_role" "iam_sa_role" {
  name               = "role-sa-${local.this_service_name}-${var.environment}"
  path               = "/"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowEksAuthToAssumeRoleForPodIdentity",
            "Effect": "Allow",
            "Principal": {
                "Service": "pods.eks.amazonaws.com"
            },
            "Action": [
                "sts:AssumeRole",
                "sts:TagSession"
            ]
        }
    ]
}
EOF
}

####### IAM policy for SA 

resource "aws_iam_policy" "dynamodb_iam_policy" {
  name        = "policy-dynamodb-${local.this_service_name}-${var.environment}"
  path        = "/"
  description = "policy for k8 service account"
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
          "${aws_dynamodb_table.dynamo_table.arn}/index/by_email",
          "${aws_dynamodb_table.dynamo_table.arn}/index/by_username"
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "att_policy_role1" {
  policy_arn = aws_iam_policy.dynamodb_iam_policy.arn
  role       = aws_iam_role.iam_sa_role.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = local.cluster_name
  namespace       = var.environment
  service_account = "sa-${local.this_service_name}-service"
  role_arn        = aws_iam_role.iam_sa_role.arn
}

#######################
####  ECR Repo

module "ecr_registry" {
  source       = "../../modules/ecr"
  depends_on   = [aws_iam_role.iam_sa_role]
  name         = "${local.this_service_name}-service"
  region       = var.region
  project_name = var.project_name
  environment  = var.environment
  account_id   = data.aws_caller_identity.id_account.id
  service_role = aws_iam_role.iam_sa_role.arn
}

###########################
##### K8 Manifests 

# Add Kustomization to flux
resource "github_repository_file" "kustomization" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/${local.this_service_name}-kustomize.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/kustomization/${local.this_service_name}.yaml",
    {
      ENVIRONMENT  = var.environment
      SERVICE_NAME = local.this_service_name
      SERVICE_NAME               = local.this_service_name
      ECR_REPO                   = module.ecr_registry.repo_url
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
  # lifecycle {
  #   ignore_changes = [content]
  # }
}

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
      SERVICE_PATH_HEALTH_CHECKS = "${local.this_service_name}/health" ## don't include the / at the beginning
      TIER                       = local.tier
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


##### overlays
###Patch
resource "github_repository_file" "overlays_svc_patch" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/overlays/${var.environment}/patch-deployment.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/patch-deployment.yaml",
    {
      SERVICE_NAME        = local.this_service_name
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.dynamo_table.name
      AWS_REGION          = var.region
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
## Kustomization
resource "github_repository_file" "overlays_svc_kustomization" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/overlays/${var.environment}/kustomization.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/kustomization.yaml",
    {
      SERVICE_NAME = local.this_service_name
      ECR_REPO     = module.ecr_registry.repo_url
      ENVIRONMENT  = var.environment
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


##### Network Policies
resource "github_repository_file" "network_policy" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/${local.this_service_name}.yaml",
    {
      PROJECT_NAME = var.project_name
      SERVICE_PORT = local.this_service_port
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


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
      PATH_DEPLOYMENT = "clusters/${local.cluster_name}/bootstrap/flux-system/${local.this_service_name}-kustomize.yaml"
      
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

#services/${local.this_service_name}-service/overlays/${var.environment}/kustomization.yaml"