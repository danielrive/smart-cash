
locals {
  eks_cluster_name   = var.cluster_name
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

resource "aws_iam_role_policy_attachment" "AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_iam_role.name
}

resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly-EKS" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
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
    authentication_mode = "API"
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
                "AWS": "arn:aws:iam::${var.account_number}:root"
            },
            "Action": "sts:AssumeRole",
            "Condition": {}
        }
    ]
}
EOF
}

resource "aws_eks_access_policy_association" "eks_admin" {
  depends_on = [aws_eks_cluster.kube_cluster]
  cluster_name  = aws_eks_cluster.kube_cluster.name
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy"
  principal_arn = aws_iam_user.eks_admin_iam_role.arn

  access_scope {
    type       = "namespace"
    namespaces = ["*"]
  }
}



################################
#####  EKS worker node role ####
################################

/*
Nodes must have a role that allows to make calls to AWS API, the role is associate to a instance profile
that is attached to EC2 instance
*/

resource "aws_iam_role" "workernodes" {
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

resource "aws_iam_role_policy_attachment" "AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.workernodes.name
}

resource "aws_iam_role_policy_attachment" "AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.workernodes.name
}

resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.workernodes.name
}



################################
#####  EKS manage node group####
################################

/*
node group managed by eks, this contains the ec2 instances that will be the worker nodes
ec2 instances has associated the node role created before

*/
resource "aws_eks_node_group" "worker-node-group" {
  cluster_name    = local.eks_cluster_name
  node_group_name = local.eks_node_group_name
  node_role_arn   = aws_iam_role.workernodes.arn
  subnet_ids      = var.subnet_ids
  ami_type        = var.AMI_for_worker_nodes
  instance_types  = var.instance_type_worker_nodes
  scaling_config {
    desired_size = var.min_instances_node_group
    max_size     = var.max_instances_node_group
    min_size     = var.min_instances_node_group
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Node Group handling.
  # Otherwise, EKS will not be able to properly delete EC2 Instances and Elastic Network Interfaces.
  depends_on = [
    aws_iam_role_policy_attachment.AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.AmazonEKS_CNI_Policy,
    aws_eks_cluster.kube_cluster,
    aws_iam_role_policy_attachment.AmazonEC2ContainerRegistryReadOnly
  ]
}


##############################################################################################################################
# AWS EKS VPC CNI plugin
# https://docs.aws.amazon.com/eks/latest/userguide/cni-iam-role.html
#############################################################################################################################

resource "null_resource" "vpc_cni_plugin_for_iam" {
  provisioner "local-exec" {
    command = <<EOF
      curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
      /tmp/eksctl version
      /tmp/eksctl create iamserviceaccount --name aws-node --namespace kube-system --cluster ${aws_eks_cluster.kube_cluster.name} --region ${var.region} --role-name "${aws_eks_cluster.kube_cluster.name}_AmazonEKSVPCCNIRole" --attach-policy-arn arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy --override-existing-serviceaccounts --approve
    EOF
  }
  depends_on = [
    aws_eks_cluster.kube_cluster,
    aws_eks_node_group.worker-node-group
  ]

}

########################################
# VPC CNI
#########################################

resource "aws_eks_addon" "vpc-cni" {
  cluster_name      = aws_eks_cluster.kube_cluster.name
  addon_name        = "vpc-cni"
  resolve_conflicts_on_create = "OVERWRITE"

  configuration_values = jsonencode({
    enableNetworkPolicy= "true"
  })
}
