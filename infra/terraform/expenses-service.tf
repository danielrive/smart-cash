################################################
########## Resources for expenses-service

#######################
#### DynamoDB tables

### expenses Table

resource "aws_dynamodb_table" "expenses_table" {
  name         = "expenses-table"
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
    non_key_attributes = ["expenseId","currency","date","amount"]
  }

  global_secondary_index {
    name               = "by_category"
    hash_key           = "userId"
    range_key           = "category"
    projection_type    = "ALL"
  }
  tags = {
    Name = "expenses_${var.environment}"
  }
  
}


##############################
###### IAM Role K8 SA

resource "aws_iam_role" "expenses-role" {
  name = "role-expenses-${var.environment}"
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
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-expenses-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA expenses

resource "aws_iam_policy" "dynamodb-expenses-policy" {
  name        = "policy-dynamodb-expenses-${var.environment}"
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
                    aws_dynamodb_table.expenses_table.arn,
                    "${aws_dynamodb_table.expenses_table.arn}/index/by_userId",
                    "${aws_dynamodb_table.expenses_table.arn}/index/by_category"
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-expenses-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-expenses-policy.arn
  role       = aws_iam_role.expenses-role.name
}



#############################
##### ECR Repo

module "ecr_registry_expenses_service" {
  source       = "./modules/ecr"
  name         = "expenses-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests-expenses-svc" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/expenses-service/base/${each.key}"
  content = templatefile(
    "../kubernetes/microservices-templates/${each.key}",
    {
      SERVICE_NAME = "expenses"
      SERVICE_PORT = "8282"
      ECR_REPO = module.ecr_registry_expenses_service.repo_url
      SERVICE_PATH_HEALTH_CHECKS = "/health"     
      SERVICE_PORT_HEALTH_CHECKS = "8282"
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

resource "github_repository_file" "overlays-expenses-svc" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/expenses-service/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/expenses-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "../kubernetes/expenses-service/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = "expenses"
      ECR_REPO = module.ecr_registry_expenses_service.repo_url
      ARN_ROLE_SERVICE = aws_iam_role.expenses-role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.expenses_table.name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
##### Network Policies

resource "github_repository_file" "np-expenses" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/manifests/user-service/base/network-policy.yaml"
  content = templatefile(
    "../kubernetes/network-policies/expenses.yaml"
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}