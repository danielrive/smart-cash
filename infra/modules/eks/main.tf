
locals {
  eks_node_group_name = "${var.project_name}-${var.environment}-eks-node-group"
}

####  IAM EKS Role  
# Role that will be used by the EKS cluster to make calls to aws services.

resource "aws_iam_role" "eks_iam_role" {
  name               = "role-eks-${var.cluster_name}-${var.region}"
  path               = "/"
  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Principal": {
    "Service": "eks.amazonaws.com"
   },
   "Action": "sts:AssumeRole"
  }
 ]
}
EOF
}

## Attach policy to role
resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_iam_role.name
}


### cloudwatch logs group 

resource "aws_cloudwatch_log_group" "log_groups_control_plane" {
  name              = "/aws/eks/${var.cluster_name}/cluster"
  retention_in_days = var.retention_control_plane_logs
  kms_key_id        = var.kms_arn
}

resource "aws_cloudwatch_log_group" "log_groups_workloads" {
  name              = "/aws/eks/${var.cluster_name}/workloads"
  retention_in_days = var.retention_control_plane_logs
  kms_key_id        = var.kms_arn
}


##### EKS Cluster  ####

resource "aws_eks_cluster" "kube_cluster" {
  depends_on                = [aws_cloudwatch_log_group.log_groups_control_plane]
  name                      = var.cluster_name
  role_arn                  = aws_iam_role.eks_iam_role.arn
  version                   = var.cluster_version
  enabled_cluster_log_types = var.cluster_enabled_log_types
  encryption_config {
    provider {
      key_arn = var.kms_arn
    }
    resources = ["secrets"]
  }
  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = true
  }
  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = var.private_endpoint_api
    endpoint_public_access  = var.public_endpoint_api
  }
  tags = {
    "karpenter.sh/discovery" = var.cluster_name
  }
}

#### IAM role used to manage the entire cluster, for now you need to pass one user that will be able to assume the role

resource "aws_iam_role" "eks_admin_iam_role" {
  name               = "admin-role-eks-${var.cluster_name}-${var.region}"
  path               = "/"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::${var.account_number}:user/${var.cluster_admins}"  
            },
            "Action": "sts:AssumeRole",
            "Condition": {}
        }
    ]
}
EOF
}

## Create EKS access entries
resource "aws_eks_access_entry" "eks_admin_entry" {
  depends_on    = [aws_eks_cluster.kube_cluster, aws_iam_role.eks_admin_iam_role]
  cluster_name  = aws_eks_cluster.kube_cluster.name
  principal_arn = aws_iam_role.eks_admin_iam_role.arn
  type          = "STANDARD"
}

## Assign full permissions over the entire cluster
resource "aws_eks_access_policy_association" "eks_admin" {
  depends_on    = [aws_eks_access_entry.eks_admin_entry]
  cluster_name  = aws_eks_cluster.kube_cluster.name
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
  principal_arn = aws_iam_role.eks_admin_iam_role.arn
  access_scope {
    type = "cluster"
  }
}

## Create Identity Policy eks policy for role

resource "aws_iam_policy" "eks-admin" {
  name   = "policy-admin-eks-${var.cluster_name}-${var.region}"
  path   = "/"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowClusterMgnt",
            "Effect": "Allow",
            "Action": [
                "eks:*"
            ],
            "Resource": "${aws_eks_cluster.kube_cluster.arn}"
        },
        {
          "Sid": "AllowListClusters",
            "Effect": "Allow",
            "Action": [
                "eks:DescribeCluster",
                "eks:ListClusters"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}

## Attach policy to role
resource "aws_iam_role_policy_attachment" "eks_admin_role" {
  policy_arn = aws_iam_policy.eks-admin.arn
  role       = aws_iam_role.eks_admin_iam_role.name
}


##################
## OIDC Config ###

# Configure OIDC for IRSA(IAM Roles for Service Accounts)

# Get tls certificate from EKS cluster identity issuer
data "tls_certificate" "cluster" {
  depends_on = [aws_eks_cluster.kube_cluster]
  url        = aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer
}

# To associate default OIDC provider to Kube cluster
resource "aws_iam_openid_connect_provider" "kube_cluster_oidc_provider" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cluster.certificates.0.sha1_fingerprint]
  url             = aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer
}



#####  EKS worker node role ####

resource "aws_iam_role" "worker_nodes" {
  name = "role-${local.eks_node_group_name}"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

## Attach policy to worker node role
resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.worker_nodes.name
}

resource "aws_iam_role_policy_attachment" "ecr_read_only" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.worker_nodes.name
}

resource "aws_iam_role_policy_attachment" "cni_policy_node" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.worker_nodes.name
}

resource "aws_iam_role_policy_attachment" "ssm_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.worker_nodes.name
}

################################
#####  EKS manage node group ###
################################

// Launch configuration for Node Group

resource "aws_launch_template" "node_group" {
  name          = "template-eks-${local.eks_node_group_name}"
  instance_type = var.instance_type_worker_nodes
  key_name      = var.key_pair_name
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size = var.storage_nodes
    }
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    instance_metadata_tags      = "enabled"
  }
  monitoring {
    enabled = true
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "template-eks-${local.eks_node_group_name}"
    }
  }
}

/*
node group managed by eks, this contains the ec2 instances that will be the worker nodes
ec2 instances has associated the node role created before
*/
resource "aws_eks_node_group" "worker-node-group" {
  cluster_name    = var.cluster_name
  node_group_name = local.eks_node_group_name
  node_role_arn   = aws_iam_role.worker_nodes.arn
  subnet_ids      = var.subnet_ids
  ami_type        = var.AMI_for_worker_nodes
  update_config {
    max_unavailable = 1
  }
  scaling_config {
    desired_size = var.min_instances_node_group
    max_size     = var.max_instances_node_group
    min_size     = var.min_instances_node_group
  }
  launch_template {
    id      = aws_launch_template.node_group.id
    version = aws_launch_template.node_group.latest_version
  }
  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_eks_cluster.kube_cluster,
    aws_iam_role_policy_attachment.ecr_read_only,
    aws_launch_template.node_group
  ]
}

### VPC CNI  ###

// IAM role for CNI add-on
resource "aws_iam_role" "vpc_cni_role" {
  name               = "vpc-cni-${var.cluster_name}-${var.region}"
  path               = "/"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::${var.account_number}:oidc-provider/${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}:aud": "sts.amazonaws.com",
                    "${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}:sub": "system:serviceaccount:kube-system:aws-node"
                }
            }
        }
    ]
}
EOF
}

## Attach policy to vpc cni role
resource "aws_iam_role_policy_attachment" "cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.vpc_cni_role.name
}

## Install CNI add-on
resource "aws_eks_addon" "vpc-cni" {
  cluster_name                = aws_eks_cluster.kube_cluster.name
  addon_name                  = "vpc-cni"
  addon_version               = var.vpc_cni_version
  service_account_role_arn    = aws_iam_role.vpc_cni_role.arn
  resolve_conflicts_on_update = "OVERWRITE"
  configuration_values = jsonencode({
    enableNetworkPolicy = "true"
  })
}


################
### EBS CSI  ###

// IAM role for CSI add-on
resource "aws_iam_role" "ebs_csi_role" {
  name               = "ebs-cni-role-${var.cluster_name}-${var.region}"
  path               = "/"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::${var.account_number}:oidc-provider/${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}:aud": "sts.amazonaws.com",
                    "${replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")}:sub": "system:serviceaccount:kube-system:ebs-csi-controller-sa"
                }
            }
        }
    ]
}
EOF
}

## Attach policy to vpc cni role
resource "aws_iam_role_policy_attachment" "csi_policy" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
  role       = aws_iam_role.ebs_csi_role.name
}

/*
## Install EBS add-on
resource "aws_eks_addon" "ebs_csi" {
  cluster_name                = aws_eks_cluster.kube_cluster.name
  addon_name                  = "aws-ebs-csi-driver"
  addon_version               = var.ebs_csi_version
  service_account_role_arn    = aws_iam_role.ebs_csi_role.arn
  resolve_conflicts_on_update = "OVERWRITE"
}
*/


#####################
### Pod Identity  ###
#####################

## Install EBS add-on
resource "aws_eks_addon" "pod_identity" {
  cluster_name                = aws_eks_cluster.kube_cluster.name
  addon_name                  = "eks-pod-identity-agent"
  addon_version               = var.pod_identity_version
  resolve_conflicts_on_update = "OVERWRITE"
}