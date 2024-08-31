locals {
  brach_gitops_repo        = var.environment
  cluster_name             = "${var.project_name}-${var.environment}"
  domain_name              = "danielrive.site"
}


### Common ArgoApps 

resource "github_repository_file" "common_argo_apps" {
  for_each   = fileset("./k8-manifests/argo-apps", "*.yaml")
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/${each.key}"
  content = templatefile(
    "./k8-manifests/argo-apps/${each.key}",
    {
      ## Common variables for manifests
      ENVIRONMENT           = var.environment
      REPO_URL = data.github_repository.gh_gitops.http_clone_url
      GITOPS_PATH_COMMON = "clusters/${local.cluster_name}/common"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

###########################
#### Common resources

resource "github_repository_file" "common_resources" {
  depends_on = [github_repository_file.common_argo_apps]
  for_each   = fileset("./k8-manifests/common/", "*.yaml")
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/common/${each.key}"
  content = templatefile(
    "./k8-manifests/common/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION  = var.region
      ENVIRONMENT = var.environment
      PROJECT     = var.project_name
      DOMAIN_NAME = local.domain_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


#################################
##### OPA constraints(policies)

resource "github_repository_file" "opa_constraints" {
  depends_on = [github_repository_file.common_resources]
  repository = data.github_repository.gh_gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/common/opa-constraints.yaml"
  content = templatefile(
    "./k8-manifests/opa-policies/constraints.yaml",
    {
      ECR_REGISTRY = "${data.aws_caller_identity.id_account.id}.dkr.ecr.${var.region}.amazonaws.com"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
