resource "aws_cloudformation_stack" "network" {
  name = "karpenter-core-${var.cluster_name}"

  parameters = {
    ClusterName = var.cluster_name
  }
  template_body = file("${path.module}/cloudformation.yaml")
  capabilities = ["CAPABILITY_NAMED_IAM"]

}

// Role for Karpenter Service Account

resource "aws_iam_role" "iam_sa_role" {
  name               = "role-sa-karpenter-${var.environment}"
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

resource "aws_iam_role_policy_attachment" "att_policy_role1" {
  policy_arn = "arn:aws:iam::${var.account_number}:role/KarpenterControllerPolicy-${cluster_name}"
  role       = aws_iam_role.iam_sa_role.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = var.cluster_name
  namespace       = var.environment
  service_account = "karpenter"
  role_arn        = aws_iam_role.iam_sa_role.arn
}