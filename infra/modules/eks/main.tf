
locals {
  eks_cluster_name    = var.cluster_name
  eks_node_group_name = "${var.project_name}-${var.environment}-eks-node-group"
}

################################
#####     IAM EKS Role     #####
################################

/*
Role that will be used by the EKS cluster to make calls to aws services like ec2 instances, tag ec2 instances.
create security groups, etcs
This role must be created before the cluster creation 
*/

resource "aws_iam_role" "eks_iam_role" {
  name = "role-eks-${local.eks_cluster_name}-${var.region}"

  path = "/"

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

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_iam_role.name
}

#############################
### cloudwatch logs group ###
#############################

resource "aws_cloudwatch_log_group" "log_groups_eks" {
  name              = "/aws/eks/${local.eks_cluster_name}/cluster"
  retention_in_days = var.retention_control_plane_logs
  kms_key_id        = var.kms_arn
}

################################
#####      EKS Cluster     #####
################################

resource "aws_eks_cluster" "kube_cluster" {
  depends_on = [aws_cloudwatch_log_group.log_groups_eks]
  name       = local.eks_cluster_name
  role_arn   = aws_iam_role.eks_iam_role.arn
  version    = var.cluster_version
  encryption_config {
    provider {
      key_arn = var.kms_arn
    }
    resources = ["secrets"]
  }
  enabled_cluster_log_types = var.cluster_enabled_log_types
  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = true
  }
  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = var.private_endpoint_api
    endpoint_public_access  = var.public_endpoint_api
  }
}

// Enable access to eks cluster to iam role for cluster management

################################
#### IAM ROLE CLUSTER ADMIN 

resource "aws_iam_role" "eks_admin_iam_role" {
  name = "admin-role-eks-${local.eks_cluster_name}-${var.region}"

  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::${var.account_number}:user/daniel.rivera"  
            },
            "Action": "sts:AssumeRole",
            "Condition": {}
        }
    ]
}
EOF
}

resource "aws_eks_access_entry" "eks_admin_entry" {
  depends_on    = [aws_eks_cluster.kube_cluster, aws_iam_role.eks_admin_iam_role]
  cluster_name  = aws_eks_cluster.kube_cluster.name
  principal_arn = aws_iam_role.eks_admin_iam_role.arn
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "eks_admin" {
  depends_on    = [aws_eks_access_entry.eks_admin_entry]
  cluster_name  = aws_eks_cluster.kube_cluster.name
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy"
  principal_arn = aws_iam_role.eks_admin_iam_role.arn

  access_scope {
    type       = "namespace"
    namespaces = ["*"]
  }
}



/// Configure OIDC for IRSA(IAM Roles for Service Accounts)

#########################
## OIDC Config
#########################
# Get tls certificate from EKS cluster identity issuer

data "tls_certificate" "cluster" {
  url = aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer
  depends_on = [
    aws_eks_cluster.kube_cluster
  ]
}

# To associate default OIDC provider to Kube cluster

resource "aws_iam_openid_connect_provider" "kube_cluster_oidc_provider" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cluster.certificates.0.sha1_fingerprint]
  url             = aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer
}




################################
#####  EKS worker node role ####
################################

/*
Nodes must have a role that allows to make calls to AWS API, the role is associate to a instance profile
that is attached to EC2 instance
*/

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

resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.worker_nodes.name
}

resource "aws_iam_role_policy_attachment" "ecr_read_only" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.worker_nodes.name
}



################################
#####  EKS manage node group####
################################

// Launch configuration for Node Group

resource "aws_launch_template" "node_group" {
  name          = "template-eks-${local.eks_node_group_name}"
  image_id      = var.AMI_for_worker_nodes
  instance_type = var.instance_type_worker_nodes
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
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
  cluster_name    = local.eks_cluster_name
  node_group_name = local.eks_node_group_name
  node_role_arn   = aws_iam_role.worker_nodes.arn
  subnet_ids      = var.subnet_ids
  #update_config {
  #  max_unavailable = 1
  #}
  scaling_config {
    desired_size = var.min_instances_node_group
    max_size     = var.max_instances_node_group
    min_size     = var.min_instances_node_group
  }
  launch_template {
    id      = aws_launch_template.node_group.id
    version = aws_launch_template.node_group.latest_version
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Node Group handling.
  # Otherwise, EKS will not be able to properly delete EC2 Instances and Elastic Network Interfaces.
  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_eks_cluster.kube_cluster,
    aws_iam_role_policy_attachment.ecr_read_only,
    aws_launch_template.node_group
  ]
}


########################################
# VPC CNI
#########################################

// IAM role for CNI add-on

resource "aws_iam_role" "vpc_cni_role" {
  name = "vpc-cni-s-${local.eks_cluster_name}-${var.region}"

  path = "/"

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

resource "aws_iam_role_policy_attachment" "cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.vpc_cni_role.name
}

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