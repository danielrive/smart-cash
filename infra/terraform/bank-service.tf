################################################
########## Resources for bank-service


#######################
#### DynamoDB tables

### bank Table

resource "aws_dynamodb_table" "transactions_table" {
  name         = "transactions-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "transactionId"

  attribute {
    name = "transactionId"
    type = "S"
  }


  tags = {
    ENVIRONMENT = "${var.environment}"
  }
  
}


##############################
###### IAM Role K8 SA

resource "aws_iam_role" "bank-role" {
  name = "role-bank-${var.environment}"
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
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-bank-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA bank

resource "aws_iam_policy" "dynamodb-bank-policy" {
  name        = "policy-dynamodb-bank-${var.environment}"
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
                    aws_dynamodb_table.transactions_table.arn,
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-bank-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-bank-policy.arn
  role       = aws_iam_role.bank-role.name
}



#############################
##### ECR Repo

module "ecr_registry_bank_service" {
  source       = "./modules/ecr"
  name         = "bank-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests-bank-svc" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/bank-service/base/${each.key}"
  content = templatefile(
    "../kubernetes/microservices-templates/${each.key}",
    {
      SERVICE_NAME = "bank"
      SERVICE_PORT = "8282"
      SERVICE_PATH_HEALTH_CHECKS = "/health"     
      SERVICE_PORT_HEALTH_CHECKS = "8282"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



###########################
##### overlays

resource "github_repository_file" "overlays-bank-svc" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  for_each            = fileset("../kubernetes/bank-service/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/bank-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "../kubernetes/bank-service/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = "bank"
      ECR_REPO = module.ecr_registry_bank_service.repo_url
      ARN_ROLE_SERVICE = aws_iam_role.bank-role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.transactions_table.name
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

resource "github_repository_file" "np-bank" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations-bootstrap]
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/bank-service/base/network-policy.yaml"
  content = templatefile(
    "../kubernetes/network-policies/bank.yaml",{
      PROJECT_NAME  = var.project_name
    })
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}