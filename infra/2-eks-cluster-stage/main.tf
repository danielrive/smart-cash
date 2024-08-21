locals {
  brach_gitops_repo               = var.environment
  path_app_bootstrap              = "./k8-manifests/bootstrap/kustomizations"
  path_tf_repo_flux_sources       = "./k8-manifests/bootstrap/flux-sources"
  path_tf_repo_base          = "./k8-manifests/core"
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


############################
#####  ArgoCD Bootstrap 

// Get kubeconfig GH runner to run HELM
resource "null_resource" "install_argo" {
  depends_on = [module.eks_cluster]
  provisioner "local-exec" {
    command = <<EOF
    echo "---> install helm"
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    chmod 700 get_helm.sh
    ./get_helm.sh
    echo "---> get kubeconfig"
    aws eks update-kubeconfig --name ${local.cluster_name} --region ${var.region}
    echo "---> install argo"
    helm repo add argo https://argoproj.github.io/argo-helm
    helm repo update
    helm install argocd argo/argo-cd --namespace argocd --create-namespace  -f ./k8-manifests/helm-argo-installation/argocd.yaml
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }
}

resource "null_resource" "bootstrap_argo" {
  depends_on = [module.eks_cluster]
  provisioner "local-exec" {
    command = <<EOF
    echo "---> install argocli"
    curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
    sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
    rm argocd-linux-amd64
    echo "---> get kubeconfig"
    aws eks update-kubeconfig --name ${local.cluster_name} --region ${var.region}
    echo "---> add github repo"
    argocd repo add ${data.github_repository.gh_gitops.http_clone_url} --password $GITHUB_TOKEN --username argobot
    echo "---> add main app"
    argocd app create main-app --repo ${data.github_repository.gh_gitops.http_clone_url} \ 
    --revision ${var.environment}
    --path cluster/${local.cluster_name}/bootstrap \
    --dest-namespace argocd --dest-server ${module.eks_cluster.cluster_endpoint} \
    --directory-recurse
    EOF
  }
  triggers = {
    always_run = timestamp() # this will always run
  }
}

// Bootstrap First main app



/*
# configure Private Repo
resource "argocd_repository" "gh_gitops" {
  depends_on      = [helm_release.install_argo]
  repo            = data.github_repository.gh_gitops.http_clone_url
  username        = "argobot"
  password        = var.gh_token
  insecure        = false
}

# Argocd app for base manifest
resource "argocd_application" "base" {
  depends_on      = [argocd_repository.gh_gitops]
  metadata {
    name      = "base"
    namespace = "argocd"
    labels = {
      project = var.project_name
    }
  }
  cascade = false # disable cascading deletion
  wait    = true
  spec {
    project = "default"
    destination {
      server    = module.eks_cluster.cluster_endpoint
      namespace = var.environment
    }
    source {
      repo_url        = data.github_repository.gh_gitops.http_clone_url
      path            = "cluster/${local.cluster_name}/base"
      target_revision = var.environment == "prod" ? "main" : var.environment
    }
    sync_policy {
      automated {
        prune       = true
        self_heal   = true
      }
    }
  }
}

## Push the base manifests 

##### Base resources
resource "github_repository_file" "base_resources" {
  depends_on = [module.eks_cluster, null_resource.bootstrap-flux]
  for_each   = fileset("./k8-manifests/base", "*.yaml")
  repository = data.github_repository.flux-gitops.name
  branch     = local.brach_gitops_repo
  file       = "clusters/${local.cluster_name}/base/${each.key}"
  content = templatefile(
    "${local.path_tf_repo_flux_core}/${each.key}",
    {
      ## Common variables for manifests
      AWS_REGION            = var.region
      ENVIRONMENT           = var.environment
      PROJECT               = var.project_name
      ARN_CERT_MANAGER_ROLE = "arn:aws:iam::12345678910:role/cert-manager-us-west-2"
    }
  )
  commit_message      = "Managed by Terraform"
  commit_author       = "From terraform"
  commit_email        = "gitops@smartcash.com"
  overwrite_on_create = true
}

*/