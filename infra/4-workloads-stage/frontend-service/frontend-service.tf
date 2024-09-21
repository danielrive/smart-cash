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
  name = "role-sa-${local.this_service_name}-${var.environment}"
  path = "/"
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

#############################
##### ECR Repo

module "ecr_registry" {
  source       = "../../modules/ecr"
  name         = "frontend-service"
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
      ENVIRONMENT               = var.environment
      SERVICE_NAME              = local.this_service_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



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
      SERVICE_PATH_HEALTH_CHECKS = "/index.html"
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

###Patch
resource "github_repository_file" "overlays_svc_patch" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/${local.this_service_name}-service/overlays/${var.environment}/patch-deployment.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/patch-deployment.yaml",
    {
      SERVICE_NAME        = local.this_service_name
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
      SERVICE_NAME        = local.this_service_name
      ECR_REPO            = module.ecr_registry.repo_url
      ENVIRONMENT         = var.environment
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
  lifecycle {
    ignore_changes = [content]
  }
}


###########################
##### Network Policies

resource "github_repository_file" "network_policy" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/frontend-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/frontend.yaml", {
      PROJECT_NAME = var.project_name
      SERVICE_PORT = local.this_service_port
  })
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