locals {
  brach_gitops_repo = var.environment
  path_tf_repo_flux_kustomization = "./k8-manifests/bootstrap/kustomizations"
  path_tf_repo_flux_sources = "./k8-manifests/bootstrap/flux-sources"
  path_tf_repo_flux_core = "./k8-manifests/core"
  cluster_name = "${var.project_name}-${var.environment}"
  gh_username = "danielrive"
}


##########################
####### EKS Cluster


module "eks_cluster" {
  source                       = "../modules/eks"
  environment                  = var.environment
  region                       = var.region
  cluster_name                 = local.cluster_name
  project_name                 = var.project_name
  cluster_version              = "1.29"
  subnet_ids                   = data.terraform_remote_state.base.outputs.public_subnets
  retention_control_plane_logs = 7
  instance_type_worker_nodes   = var.environment == "develop" ? "t3.medium" : "t3.medium"
  AMI_for_worker_nodes         = "AL2_x86_64"
  desired_nodes                = 2
  max_instances_node_group     = 2
  min_instances_node_group     = 2
  private_endpoint_api         = true
  public_endpoint_api          = true
  kms_arn                      = data.terraform_remote_state.base.outputs.kms_eks_arn
  account_number               = data.aws_caller_identity.id_account.id
}


###################################################################
########## IAM Role for CertManager Issuer DNS01 challenge


resource "aws_iam_role" "cert-manager-iam-role" {
  name = "cert-manager-${var.region}"

  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Principal": {
        "Federated": "arn:aws:iam::${data.aws_caller_identity.id_account.id}:oidc-provider/${module.eks_cluster.cluster_oidc}"
      },
      "Condition": {
        "StringEquals": {
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:cert-manager:cert-manager"
        }
      }
    }
  ]
}
EOF

}

resource "aws_iam_policy" "cert-manager-iam-role-policy" {
  name        = "policy-cert-manager-iam-role"
  policy      = data.aws_iam_policy_document.cert-manager-issuer.json
}

resource "aws_iam_role_policy_attachment" "cert-manager-role" {
  policy_arn = aws_iam_policy.cert-manager-iam-role-policy.arn
  role       = aws_iam_role.cert-manager-iam-role.name
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
    ./bootstrap-flux.sh ${local.cluster_name}  ${var.region} ${local.gh_username} ${data.github_repository.flux-gitops.name} ${var.environment}
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