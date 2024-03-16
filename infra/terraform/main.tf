locals {
  brach_gitops_repo = "main"
  path_tf_repo_flux_kustomization = "../kubernetes/kustomizations"
  path_tf_repo_flux_sources = "../kubernetes/flux-sources"
  path_tf_repo_flux_common = "../kubernetes/common"
  cluster_name = "${var.project_name}-${var.environment}"
}

#### Netwotking Creation

module "networking" {
  source                = "./modules/networking"
  project_name          = var.project_name
  region                = var.region
  environment           = var.environment
  cidr                  = "10.100.0.0/16"
  availability_zones    = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  private_subnets       = ["10.100.0.0/22", "10.100.64.0/22", "10.100.128.0/22"]
  db_subnets            = ["10.100.4.0/22", "10.100.68.0/22", "10.100.132.0/22"]
  public_subnets        = ["10.100.32.0/22", "10.100.96.0/22", "10.100.160.0/22"]
  enable_nat_gw         = false
  create_nat_gw         = false
  single_nat_gw         = false
  enable_auto_public_ip = true
}

### KMS Key to encrypt kubernetes resources
module "kms_key_eks" {
  source              = "./modules/kms"
  region              = var.region
  environment         = var.environment
  project_name        = var.project_name
  name                = "eks"
  key_policy          = data.aws_iam_policy_document.kms_key_policy_encrypt_logs.json
  enable_key_rotation = true
}

##########################
# EKS Cluster
##########################

module "eks_cluster" {
  source                       = "./modules/eks"
  depends_on                   = [module.kms_key_eks,module.networking]
  environment                  = var.environment
  region                       = var.region
  cluster_name                 = local.cluster_name
  project_name                 = var.project_name
  cluster_version              = "1.29"
  subnet_ids                   = module.networking.main.public_subnets
  retention_control_plane_logs = 7
  instance_type_worker_nodes   = var.environment == "develop" ? ["t3.medium"] : ["t3.medium"]
  AMI_for_worker_nodes         = "AL2_x86_64"
  desired_nodes                = 1
  max_instances_node_group     = 1
  min_instances_node_group     = 1
  private_endpoint_api         = true
  public_endpoint_api          = true
  kms_arn                      = module.kms_key_eks.kms_arn
  userRoleARN                  = "arn:aws:iam::${data.aws_caller_identity.id_account.id}:role/user-mgnt-eks-cluster"
}


###############################################
#######    GitOps Configuration 
###############################################

###########################
##### Common kustomization

resource "github_repository_file" "common_kustomize" {
  depends_on          = [module.eks_cluster]
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/common-kustomize.yaml"
  content = templatefile(
    "${local.path_tf_repo_flux_kustomization}/common-kustomize.yaml",
    {
      PROJECT_NAME             = var.project_name
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}
###########################
##### HELM Sources 

resource "github_repository_file" "sources" {
  depends_on          = [module.eks_cluster]
  for_each            = fileset(local.path_tf_repo_flux_sources, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "clusters/${local.cluster_name}/${each.key}"
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
##### Common resources

resource "github_repository_file" "common_resources" {
  depends_on          = [module.eks_cluster]
  for_each            = fileset(local.path_tf_repo_flux_common, "*.yaml")
  repository          = data.github_repository.flux-gitops.name
  branch              = local.brach_gitops_repo
  file                = "$common/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_common}/${each.key}",
    {}
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}



###########################
#####


###########################
#####


###########################
#####


###########################
#####


###########################
#####

