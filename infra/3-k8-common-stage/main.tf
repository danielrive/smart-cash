locals {
  brach_gitops_repo        = var.environment
  path_tf_repo_flux_common = "./k8-manifests/common"
  cluster_name             = "${var.project_name}-${var.environment}"
  gh_username              = "danielrive"
}


###########################
##### Common resources

resource "github_repository_file" "common_resources" {
  for_each   = fileset(local.path_tf_repo_flux_common, "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/common/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_common}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION  = var.region
      ENVIRONMENT = var.environment
      PROJECT     = var.project_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

############################
##### OPA constraints(policies)

resource "github_repository_file" "opa_constraints" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/opa-policies/opa-constraints.yaml"
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