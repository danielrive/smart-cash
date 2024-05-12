locals {
  brach_gitops_repo = "main"
  path_tf_repo_flux_common = "../../../kubernetes/common"
  cluster_name = "${var.project_name}-${var.environment}"
  gh_username = "danielrive"
}



###########################
##### Common resources

resource "github_repository_file" "common_resources" {
  for_each            = fileset(local.path_tf_repo_flux_common, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/common/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_common}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION = var.region
      ENVIRONMENT = var.environment
      PROJECT = var.project_name
      ## Variables cert manager
      ARN_CERT_MANAGER_ROLE = "arn:aws:iam::12345678910:role/cert-manager-us-west-2"
      ## Variables for Grafana
      ## Variables for ingress
      
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


############################
##### OPA Constraints

resource "github_repository_file" "opa_templates" {
  for_each            = fileset("../kubernetes/opa-policies", "contraints*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/opa-policies/${each.key}"
  content = templatefile(
    "../../kubernetes/opa-policies/${each.key}",{}
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
