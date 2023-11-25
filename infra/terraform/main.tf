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
  environment                  = var.environment
  region                       = var.region
  project_name                 = var.project_name
  cluster_version              = "1.27"
  subnet_ids                   = module.networking.main.public_subnets
  retention_control_plane_logs = 7
  instance_type_worker_nodes   = var.environment == "develop" ? ["t3.medium"] : ["t3.medium"]
  AMI_for_worker_nodes         = "AL2_x86_64"
  desired_nodes                = 2
  max_instances_node_group     = 2
  min_instances_node_group     = 2
  private_endpoint_api         = true
  public_endpoint_api          = true
  kms_arn                      = module.kms_key_eks.kms_arn
  userRoleARN                  = "arn:aws:iam::${data.aws_caller_identity.id_account.id}:role/user-mgnt-eks-cluster"
  account_number               = data.aws_caller_identity.id_account.id
}

########################################
# IAM Role for CertManager Issuer DNS01 challenge
#########################################

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
          "${module.eks_cluster.cluster_oidc}:sub": "system:serviceaccount:${var.environment}:sa-cert-manager-issuer"
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
