locals {
  brach_gitops_repo        = var.environment
  path_tf_repo_flux_common = "./k8-manifests/common"
  cluster_name             = "${var.project_name}-${var.environment}"
  gh_username              = "danielrive"
  domain_name              = "danielrive.site"
}

### AWS Secrets Manager - JWT Secret

module "jwt_secret" {
  source      = "../modules/secret_manager"
  environment = var.environment
  project_name = var.project_name
  secret_name  = "smart-cash/jwt-secret/${var.environment}"
  description  = "JWT secret for service-to-service authentication in ${var.environment} environment"
}

### IAM Role for Secrets Store CSI Driver AWS Provider

module "secrets_store_csi_role" {
  depends_on      = [module.jwt_secret]
  source          = "../modules/secrets-store-csi-role"
  environment     = var.environment
  region          = var.region
  cluster_name    = local.cluster_name
  service_account = "secrets-store-csi-driver-provider-aws"
  namespace       = "kube-system"
  secret_arns     = [module.jwt_secret.secret_arn]
}


# Add Kustomization to flux
resource "github_repository_file" "kustomization" {
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/common-kustomize.yaml"
  content = templatefile(
    "./k8-manifests/kustomization/common.yaml",
    {
      CLUSTER_NAME = local.cluster_name
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
  depends_on = [module.jwt_secret, module.secrets_store_csi_role]
  for_each   = fileset(local.path_tf_repo_flux_common, "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/common/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_common}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION     = var.region
      ENVIRONMENT    = var.environment
      PROJECT        = var.project_name
      DOMAIN_NAME    = local.domain_name
      CLUSTER_NAME   = local.cluster_name
      JWT_SECRET_NAME = module.jwt_secret.secret_name

    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

# #################################
# ##### OPA constraints(policies)

# resource "github_repository_file" "opa_constraints" {
#   repository = data.github_repository.flux-gitops.name
#   branch     = local.brach_gitops_repo
#   file       = "clusters/${local.cluster_name}/opa-policies/opa-constraints.yaml"
#   content = templatefile(
#     "./k8-manifests/opa-policies/constraints.yaml",
#     {
#       ECR_REGISTRY = "${data.aws_caller_identity.id_account.id}.dkr.ecr.${var.region}.amazonaws.com"
#     }
#   )
#   commit_message      = "Managed by Terraform"
#   commit_author       = "From terraform"
#   commit_email        = "gitops@smartcash.com"
#   overwrite_on_create = true
# }