###############################
#### IAM Role for Secrets Store CSI Driver AWS Provider

resource "aws_iam_role" "pod_sa_role" {
  name               = "role-sa-${var.service_account}-${var.environment}-${var.region}"
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

###############################
#### Role Policy for Secrets Manager Access

resource "aws_iam_policy" "secrets_store_csi" {
  name = "policy-secrets-store-csi-${var.cluster_name}-${var.region}"
  path = "/"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = var.secret_arns
      }
    ]
  })
}

## Attach policy to role
resource "aws_iam_role_policy_attachment" "secrets_store_csi" {
  policy_arn = aws_iam_policy.secrets_store_csi.arn
  role       = aws_iam_role.pod_sa_role.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = var.cluster_name
  namespace       = var.namespace
  service_account = var.service_account
  role_arn        = aws_iam_role.pod_sa_role.arn
}
