locals {
  brach_gitops_repo               = var.environment
  path_tf_repo_flux_kustomization = "./k8-manifests/bootstrap/kustomizations"
  path_tf_repo_flux_sources       = "./k8-manifests/bootstrap/flux-sources"
  path_tf_repo_flux_core          = "./k8-manifests/core"
  cluster_name                    = "${var.project_name}-${var.environment}"
  gh_username                     = "danielrive"
}

#################################
### S3 Bucket for Grafana Tempo

resource "aws_s3_bucket" "grafana_tempo" {
  bucket = "${var.environment}-${var.project_name}-grafana-tempo-bucket"

  tags = {
    Name        = "${local.cluster_name}-grafana-tempo-bucket"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "grafana_tempo" {
  bucket = aws_s3_bucket.grafana_tempo.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = data.terraform_remote_state.base.outputs.kms_eks_arn
      sse_algorithm     = "aws:kms"
    }
  }
}

##########################
### IAM for SA Tempo

##############################
###### IAM Role K8 SA

resource "aws_iam_role" "iam_sa_role_tempo" {
  name               = "role-sa-grafana-tempo-${var.environment}"
  path               = "/"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowEksAuthToAssumeRoleForPodIdentity",
            "Effect": "Allow",
            "Principal": {
                "Service": "pods.eks.amazonaws.com"
            },
            "Action": [
                "sts:AssumeRole",
                "sts:TagSession"
            ]
        }
    ]
}
EOF
}

resource "aws_iam_policy" "tempo_iam_policy" {
  name        = "policy-grafana-tempo-${var.environment}"
  path        = "/"
  description = "policy for k8 service account"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:*",
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.grafana_tempo.arn,
          "${aws_s3_bucket.grafana_tempo.arn}/*"
        ]
      },
      {
        Action = [
          "kms:*",
        ]
        Effect = "Allow"
        Resource = [
          data.terraform_remote_state.base.outputs.kms_eks_arn
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "att_policy_role1" {
  policy_arn = aws_iam_policy.tempo_iam_policy.arn
  role       = aws_iam_role.iam_sa_role_tempo.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = local.cluster_name
  namespace       = "monitoring"
  service_account = "sa-grafana-tempo"
  role_arn        = aws_iam_role.iam_sa_role_tempo.arn
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
  cluster_version              = "1.31"
  subnet_ids                   = data.terraform_remote_state.base.outputs.public_subnets ## fLUX NEED INTERNET ACCESS, NAT not used to avoid costs
  private_endpoint_api         = true
  public_endpoint_api          = true
  kms_arn                      = data.terraform_remote_state.base.outputs.kms_eks_arn
  account_number               = data.aws_caller_identity.id_account.id
  vpc_cni_version              = "v1.20.4-eksbuild.1"
  ebs_csi_version              = "v1.38.1-eksbuild.1"
  pod_identity_version         = "v1.3.4-eksbuild.1"
  cluster_admins               = "daniel.rivera" # This user will be able to assume the role to manage the cluster
  retention_control_plane_logs = 7
  cluster_enabled_log_types    = ["audit", "api", "authenticator"]
  ### configs for worker nodes
  key_pair_name              = "k8-admin"
  instance_type_worker_nodes = var.environment == "develop" ? "t3.medium" : "t3.medium"
  AMI_for_worker_nodes       = "AL2_x86_64"
  desired_nodes              = 5
  max_instances_node_group   = 5
  min_instances_node_group   = 5
  storage_nodes              = 20
}

##############################
### Flux imageupdate role

module "flux_imageupdate_role" {
  depends_on      = [module.eks_cluster]
  source          = "../modules/flux-image-update-role"
  environment     = var.environment
  region          = var.region
  cluster_name    = local.cluster_name
  service_account = "image-reflector-controller"
  namespace       = "flux-system"
}

######################
### cert manager role

module "cert_manager" {
  depends_on      = [module.eks_cluster]
  source          = "../modules/cert-manager"
  environment     = var.environment
  region          = var.region
  cluster_name    = local.cluster_name
  service_account = "cert-manager"
  namespace       = "cert-manager"
}

######################
### fluentbit role

module "fuent-bit-role" {
  source         = "../modules/fluent-bit-role"
  environment    = var.environment
  region         = var.region
  cluster_name   = local.cluster_name
  cluster_oidc   = module.eks_cluster.cluster_oidc
  account_number = data.aws_caller_identity.id_account.id
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

// Install kubernetes gateway resources

resource "null_resource" "install-k8-gateways" {
  depends_on = [module.eks_cluster]
  provisioner "local-exec" {
    command = <<EOF
    ./install-k8-gateway.sh ${local.cluster_name}  ${var.region}
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

##### Core resources
resource "github_repository_file" "core_resources" {
  depends_on = [module.eks_cluster, null_resource.bootstrap-flux,github_repository_file.kustomizations]
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
      CLUSTER_NAME          = local.cluster_name
      PROJECT               = var.project_name
      ARN_CERT_MANAGER_ROLE = module.cert_manager.role_arn
      ACCOUNT_NUMBER        = data.aws_caller_identity.id_account.id
      S3_BUCKET_GRAFANA_TEMPO = aws_s3_bucket.grafana_tempo.bucket
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

