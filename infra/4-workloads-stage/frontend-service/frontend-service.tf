################################################
########## Resources for frontend-service

locals {
  this_service_name     = "frontend"
  this_service_port     = 9090
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo     = var.environment
}


#############################
##### ECR Repo

module "ecr_registry" {
  source       = "../../modules/ecr"
  name         = "frontend-service"
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
      SERVICE_PATH_HEALTH_CHECKS = "index.html" ## don't include the / at the beginning
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


###########################
##### Network Policies

resource "github_repository_file" "network_policy" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "services/frontend-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/frontend.yaml", {
      PROJECT_NAME = var.project_name
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