locals {
  brach_gitops_repo               = var.environment
  path_tf_repo_flux_kustomization = "./k8-manifests/bootstrap/kustomizations"
  path_tf_repo_flux_sources       = "./k8-manifests/bootstrap/flux-sources"
  path_tf_repo_flux_core          = "./k8-manifests/core"
  cluster_name                    = "${var.project_name}-${var.environment}"
  gh_username                     = "danielrive"
}


##########################
#### EKS Cluster

module "eks_cluster" {
  source = "../modules/eks"
  ### Control plane configs
  environment                  = var.environment
  region                       = var.region
  cluster_name                 = local.cluster_name
  project_name                 = var.project_name
  cluster_version              = "1.29"
  subnet_ids                   = data.terraform_remote_state.base.outputs.public_subnets ## fLUX NEED INTERNET ACCESS, NAT not used to avoid costs
  private_endpoint_api         = true
  public_endpoint_api          = true
  kms_arn                      = data.terraform_remote_state.base.outputs.kms_eks_arn
  account_number               = data.aws_caller_identity.id_account.id
  vpc_cni_version              = "v1.18.3-eksbuild.1"
  ebs_csi_version              = "v1.34.0-eksbuild.1"
  cluster_admins               = "daniel.rivera" # This user will be able to assume the role to manage the cluster
  retention_control_plane_logs = 7
  cluster_enabled_log_types    = ["audit", "api", "authenticator"]
  ### configs for worker nodes
  key_pair_name              = "k8-admin"
  instance_type_worker_nodes = var.environment == "develop" ? "t3.medium" : "t3.medium"
  AMI_for_worker_nodes       = "AL2_x86_64"
  desired_nodes              = 2
  max_instances_node_group   = 2
  min_instances_node_group   = 2
  storage_nodes              = 20
}


######################
### cer manager role

module "flux_imageupdate_role" {
  source = "../modules/flux-image-repo-role"
  environment = var.environment
  region = var.region
  cluster_name = local.cluster_name 
  cluster_oidc = module.eks_cluster.cluster_oidc
  account_id = data.aws_caller_identity.id_account.id
}

############################
#####  Flux Bootstrap 


### Get Kubeconfig, arguments in bash script bootstrap-flux.sh
# $1 = CLUSTER_NAME
# $2 = AWS_REGION
# $3 = GH_USER_NAME
# $4 = FLUX_REPO_NAME

resource "null_resource" "bootstrap-flux" {
  depends_on = [module.eks_cluster]
  provisioner "local-exec" {
    command = <<EOF
    ./bootstrap-flux.sh ${local.cluster_name}  ${var.region} ${local.gh_username} ${data.github_repository.flux-gitops.name} ${var.environment}
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }
}

#######################################################
#####  Patch service account for imageRepositoryRole

resource "github_repository_file" "patch_flux" {
  depends_on = [module.eks_cluster, null_resource.bootstrap-flux]
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/flux-system/kustomization.yaml"
  content = templatefile(
    "./k8-manifests/bootstrap/patches-fluxBootstrap/mainKustomization.yaml",
    {
      ARN_ROLE = module.flux_imageupdate_role.role_arn
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

### Force to update the Pod to take the changes in the SA
resource "null_resource" "restart_image_reflector" {
  depends_on = [module.eks_cluster,null_resource.bootstrap-flux,github_repository_file.patch_flux]
  provisioner "local-exec" {
    command = <<EOF
    aws eks update-kubeconfig --name ${local.cluster_name}  --region ${var.region}
    flux reconcile kustomization flux-system --with-source
    sleep 5
    kubectl rollout restart deployment image-reflector-controller -n flux-system
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }
}

###############################
####  GitOps Configuration 

### Flux kustomizations bootstrap 
resource "github_repository_file" "kustomizations" {
  depends_on = [module.eks_cluster, null_resource.bootstrap-flux]
  for_each   = fileset(local.path_tf_repo_flux_kustomization, "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_kustomization}/${each.key}",
    {
      ENVIRONMENT  = var.environment
      CLUSTER_NAME = local.cluster_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}


##### Flux Sources 
resource "github_repository_file" "sources" {
  depends_on = [module.eks_cluster, github_repository_file.kustomizations]
  for_each   = fileset(local.path_tf_repo_flux_sources, "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/bootstrap/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_sources}/${each.key}",
    {}
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

######################
### cer manager role

module "cert_manager" {
  source = "../modules/cert-manager"
  environment = var.environment
  region = var.region
  cluster_name = local.cluster_name 
  cluster_oidc = module.eks_cluster.cluster_oidc
  account_id = data.aws_caller_identity.id_account.id
}

##### Core resources
resource "github_repository_file" "core_resources" {
  depends_on = [module.eks_cluster, null_resource.bootstrap-flux]
  for_each   = fileset(local.path_tf_repo_flux_core, "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/core/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_core}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION            = var.region
      ENVIRONMENT           = var.environment
      PROJECT               = var.project_name
      ARN_CERT_MANAGER_ROLE = module.cert_manager.role_arn
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}