
locals {
  eksClusterName   = "${var.project_name}-${var.environment}"
  eksNodeGroupName = "${var.project_name}-${var.environment}-eks-node-group"
}

################################
#####     IAM EKS Role     #####
################################

/*
Role that will be used by the EKS cluster to make calls to aws services like ec2 instances, tag ec2 instances.
create security groups, etcs
This role must be created before the cluster creation 
*/

resource "aws_iam_role" "eks-iam-role" {
  name = "role-eks-${local.eksClusterName}-${var.region}"

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
  role       = aws_iam_role.eks-iam-role.name
}

resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly-EKS" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks-iam-role.name
}

#############################
### cloudwatch logs group ###
#############################

resource "aws_cloudwatch_log_group" "log_groups_eks" {
  name              = "/aws/eks/${local.eksClusterName}/cluster"
  retention_in_days = var.retention_control_plane_logs
  kms_key_id        = var.kms_arn
}

################################
#####      EKS Cluster     #####
################################

resource "aws_eks_cluster" "kube_cluster" {
  depends_on = [aws_cloudwatch_log_group.log_groups_eks]
  name       = local.eksClusterName
  role_arn   = aws_iam_role.eks-iam-role.arn
  version    = var.cluster_version
  encryption_config {
    provider {
      key_arn = var.kms_arn
    }
    resources = ["secrets"]
  }
  enabled_cluster_log_types = var.cluster_enabled_log_types
  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = var.private_endpoint_api
    endpoint_public_access  = var.public_endpoint_api
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
  name = "role-${local.eksNodeGroupName}"
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

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilderECRContainerBuilds" {
  policy_arn = "arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
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
  cluster_name    = local.eksClusterName
  node_group_name = local.eksNodeGroupName
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

## OIDC Config
## https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html
#######################################
# Get tls certificate from EKS cluster identity issuer
#######################################

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


### Add IAM custom acces role to cluster

resource "null_resource" "iam-role-cluster-access" {
  provisioner "local-exec" {
    command = <<EOF
      curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
      /tmp/eksctl version
      /tmp/eksctl create iamidentitymapping --cluster ${local.eksClusterName} --region=${var.region} --arn ${var.userRoleARN} --group system:masters --username "AWSAdministratorAccess:{{SessionName}}"
    EOF
  }
  depends_on = [
    aws_eks_cluster.kube_cluster,
    aws_eks_node_group.worker-node-group
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
  resolve_conflicts = "OVERWRITE"
}

/*

########################################
# NGINX INGRESS and NLB
#########################################

resource "null_resource" "vpc_cni_plugin_for_iam" {
  provisioner "local-exec" {
    command = <<EOF
      
    EOF
  }
  depends_on = [
    aws_eks_cluster.kube_cluster,
    aws_eks_node_group.worker-node-group
  ]

}
*/