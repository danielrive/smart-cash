################################################
########## Resources for User-service
locals {
  this_service_name = "user"
  this_service_port = 8181
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
  source       = "./modules/ecr"
  name         = "user-service"
  project_name = var.project_name
  environment  = var.environment
}

###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/user-service/base/${each.key}"
  content = templatefile(
    "../kubernetes/microservices-templates/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      SERVICE_PORT = local.this_service_port
      ECR_REPO = module.ecr_registry_user_service.repo_url
      SERVICE_PATH_HEALTH_CHECKS = "/health"     
      AWS_REGION  = var.region
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
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/user-service/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/user-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "../kubernetes/user-service/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      ECR_REPO = module.ecr_registry_user_service.repo_url
      ARN_ROLE_SERVICE = aws_iam_role.user-role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.user_table.name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
##### Network Policies

resource "github_repository_file" "np-user-to-expense" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/user-service/base/network-policy.yaml"
  content = templatefile(
    "../kubernetes/network-policies/user-to-expenses.yaml",
    {
      FROM_SVC_NAME = local.this_service_name
      TO_SVC_NAME   = "expenses"
      PROJECT_NAME  = var.project_name
      TO_SVC_PORT   = "8282"

    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}