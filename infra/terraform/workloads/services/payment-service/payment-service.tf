################################################
########## Resources for payment-service

locals {
  path_tf_repo_services = "../../../../kubernetes/services"
  brach_gitops_repo = "main"
}

#######################
#### DynamoDB tables

### payment Table

resource "aws_dynamodb_table" "payment_table" {
  name         = "payment-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "orderId"

  attribute {
    name = "orderId"
    type = "S"
  }

  tags = {
    ENVIRONMENT = "${var.environment}"
  }
  
}


##############################
###### IAM Role K8 SA

resource "aws_iam_role" "payment-role" {
  name = "role-payment-${var.environment}"
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
          "${data.terraform_remote_state.eks.outputs.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-payment-service"
        }
      }
    }
  ]
})
}

####### IAM policy for SA payment

resource "aws_iam_policy" "dynamodb-payment-policy" {
  name        = "policy-dynamodb-payment-${var.environment}"
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
                    aws_dynamodb_table.payment_table.arn
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attachment-payment-policy-role1" {
  policy_arn = aws_iam_policy.dynamodb-payment-policy.arn
  role       = aws_iam_role.payment-role.name
}



#############################
##### ECR Repo

module "ecr_registry_payment_service" {
  source       = "../../../modules/ecr"
  name         = "payment-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests-payment-svc" {
  for_each            = fileset("../../../../kubernetes/microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/payment-service/base/${each.key}"
  content = templatefile(
    "../../../../kubernetes/microservices-templates/${each.key}",
    {
      SERVICE_NAME = "payment"
      SERVICE_PORT = "8383"
      SERVICE_PATH_HEALTH_CHECKS = "/health"     
      SERVICE_PORT_HEALTH_CHECKS = "8383"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



###########################
##### overlays

resource "github_repository_file" "overlays-payment-svc" {
  for_each            = fileset("${local.path_tf_repo_services}/payment-service/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/payment-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/payment-service/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = "payment"
      ECR_REPO = module.ecr_registry_payment_service.repo_url
      ARN_ROLE_SERVICE = aws_iam_role.payment-role.arn
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.payment_table.name
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

resource "github_repository_file" "np-payment" {
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "manifests/payment-service/base/network-policy.yaml"
  content = templatefile(
    "../../../../kubernetes/network-policies/payment.yaml",{
      PROJECT_NAME  = var.project_name
    })
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
