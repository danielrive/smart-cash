locals {
  brach_gitops_repo = var.environment
  path_tf_repo_flux_kustomization = "../../kubernetes/bootstrap/kustomizations"
  path_tf_repo_flux_sources = "../../kubernetes/bootstrap/flux-sources"
  path_tf_repo_flux_core = "../../kubernetes/core"
  path_tf_repo_flux_common = "../../kubernetes/common"
  cluster_name = "${var.project_name}-${var.environment}"
  gh_username = "danielrive"
}



###############################################
#######    Flux Bootstrap 


#### Get Kubeconfig
  # $1 = CLUSTER_NAME
  # $2 = AWS_REGION
  # $3 = GH_USER_NAME
  # $4 = FLUX_REPO_NAME
resource "null_resource" "bootstrap-flux" {
  depends_on          = [module.eks_cluster]
  provisioner "local-exec" {
    command = <<EOF
    ../scripts/bootstrap-flux.sh ${local.cluster_name}  ${var.region} ${local.gh_username} ${data.github_repository.flux-gitops.name} ${var.environment}
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }

}

###############################################
#######    GitOps Configuration 
###############################################


################################################
##### Flux kustomizations bootstrap /kubernetes/bootstrap
resource "github_repository_file" "kustomizations" {
  depends_on          = [module.eks_cluster,null_resource.bootstrap-flux]
  for_each            = fileset(local.path_tf_repo_flux_kustomization, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/bootstrap/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_kustomization}/${each.key}",
    {
      ENVIRONMENT = var.environment
      CLUSTER_NAME = local.cluster_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
##### Flux Sources 

resource "github_repository_file" "sources" {
  depends_on          = [module.eks_cluster,github_repository_file.kustomizations]
  for_each            = fileset(local.path_tf_repo_flux_sources, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/bootstrap/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_sources}/${each.key}",
    {}
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


###########################
##### Core resources

resource "github_repository_file" "core_resources" {
  depends_on          = [module.eks_cluster,null_resource.bootstrap-flux]
  for_each            = fileset(local.path_tf_repo_flux_core, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/core/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_core}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION = var.region
      ENVIRONMENT = var.environment
      PROJECT = var.project_name    
      ARN_CERT_MANAGER_ROLE = "arn:aws:iam::12345678910:role/cert-manager-us-west-2"    
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

###########################
##### Common resources

resource "github_repository_file" "common_resources" {
  depends_on          = [module.eks_cluster,github_repository_file.core_resources]
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
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

############################
##### OPA constraints

resource "github_repository_file" "opa_constraints" {
  depends_on          = [module.eks_cluster,github_repository_file.common_resources]
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/opa-policies/opa-constraints.yaml"
  content = templatefile(
    "../../kubernetes/opa-policies/constraints.yaml",
    {
      ECR_REGISTRY= "${data.aws_caller_identity.id_account.id}.dkr.ecr.${var.region}.amazonaws.com"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}