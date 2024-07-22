################################################
########## Resources for payment-service

locals {
  this_service_name = "frontend"
  this_service_port = 8080
  path_tf_repo_services = "./k8-manifests"
  brach_gitops_repo = var.environment
}


#############################
##### ECR Repo

module "ecr_registry_payment_service" {
  source       = "../../modules/ecr"
  name         = "frontend-service"
  project_name = var.project_name
  environment  = var.environment
}


###########################
##### K8 Manifests 

###########################
##### Base manifests

resource "github_repository_file" "base-manifests-payment-svc" {
  for_each            = fileset("../microservices-templates", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "services/user-service/base/${each.key}"
  content = templatefile(
    "../microservices-templates/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      SERVICE_PORT = local.this_service_port
      SERVICE_PATH_HEALTH_CHECKS = "index.html"      ## don't include the / at the beginning
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
  for_each            = fileset("${local.path_tf_repo_services}/overlays/${var.environment}", "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "services/user-service/overlays/${var.environment}/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_services}/overlays/${var.environment}/${each.key}",
    {
      SERVICE_NAME = local.this_service_name
      ECR_REPO = module.ecr_registry_user_service.repo_url
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
  file                = "services/payment-service/base/network-policy.yaml"
  content = templatefile(
    "${local.path_tf_repo_services}/network-policies/frontend.yaml",{
      PROJECT_NAME  = var.project_name
    })
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}