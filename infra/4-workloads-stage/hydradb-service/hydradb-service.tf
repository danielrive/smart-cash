locals {
  this_service_name     = "hydradb"
  this_service_port     = 8989
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo     = var.environment
  cluster_name          = "${var.project_name}-${var.environment}"
  tier                  = "backend"
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

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:PutItem",
          "dynamodb:DescribeTable",
          "dynamodb:UpdateItem"
        ]
        Effect   = "Allow"
        Resource = ["arn:aws:dynamodb:${var.region}:${data.aws_caller_identity.id_account.id}:table/expenses-table", "arn:aws:dynamodb:${var.region}:${data.aws_caller_identity.id_account.id}:table/bank-table", "arn:aws:dynamodb:${var.region}:${data.aws_caller_identity.id_account.id}:table/user-table"]
      },
    ],
    Statement = [
      {
        Action = [
          "s3:GetObject",
        ]
        Effect   = "Allow"
        Resource = ["arn:aws:s3:::smart-cash-fake-data/*"]
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

#############################
##### ECR Repo

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
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

#####  manifests
resource "github_repository_file" "manifests" {
  for_each   = fileset("./k8-manifests/", "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/${each.key}"
  content = templatefile(
    "./k8-manifests/${each.key}",
    {
      SERVICE_NAME    = local.this_service_name
      SERVICE_PORT    = local.this_service_port
      TIER            = local.tier
      AWS_REGION      = var.region
      ECR_REPO        = module.ecr_registry.repo_url
      ENVIRONMENT     = var.environment
      PATH_DEPLOYMENT = "services/${local.this_service_name}-service/kustomization.yaml"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
