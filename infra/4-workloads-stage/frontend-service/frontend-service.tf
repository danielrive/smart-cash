locals {
  this_service_name     = "frontend"
  this_service_port     = 9090
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo     = var.environment
  cluster_name                    = "${var.project_name}-${var.environment}"
  tier = "frontend"
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

#############################
##### ECR Repo

module "ecr_registry" {
  source       = "../../modules/ecr"
  depends_on   = [aws_iam_role.iam_sa_role]
  name         = "frontend-service"
  region       = var.region
  project_name = var.project_name
  environment  = var.environment
  account_id   = data.aws_caller_identity.id_account.id
  service_role = aws_iam_role.iam_sa_role.arn
}


###########################
##### K8 Manifests 

##  Argo needs a existing folder in the GitOps Repo, wi will push a random .txt file to create the path
##### Base resources
resource "github_repository_file" "create_init_path" {
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/services/${local.this_service_name}-service/overlays/${var.environment}/init.txt"
  content = templatefile(
    "./../init.txt",
    {}
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

# Create Argoapp
resource "github_repository_file" "argo_app" {
  depends_on = [github_repository_file.create_init_path]
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/${local.this_service_name}.yaml"
  content = templatefile(
    "./k8-manifests/argo-apps/${local.this_service_name}.yaml",
    {
      ENVIRONMENT               = var.environment
      SERVICE_NAME              = local.this_service_name
      REPO_URL                  = data.github_repository.gh_gitops.http_clone_url
      GITOPS_PATH        = "clusters/${local.cluster_name}/services/${local.this_service_name}-service/overlays/${var.environment}"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


##### Base manifests
resource "github_repository_file" "base_manifests" {
  for_each   = fileset("../microservices-templates", "*.yaml")
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/services/${local.this_service_name}-service/base/${each.key}"
  content = templatefile(
    "../microservices-templates/${each.key}",
    {
      SERVICE_NAME               = local.this_service_name
      SERVICE_PORT               = local.this_service_port
      SERVICE_PATH_HEALTH_CHECKS = "health" ## don't include the / at the beginning
      TIER                       = local.tier
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
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/services/${local.this_service_name}-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      ECR_REPO     = module.ecr_registry.repo_url
      AWS_REGION   = var.region
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
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/services/${local.this_service_name}-service/base/network-policy.yaml"
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