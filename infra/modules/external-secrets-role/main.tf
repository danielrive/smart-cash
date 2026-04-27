###############################
#### IAM Role for External Secrets Operator

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

resource "aws_iam_policy" "external_secrets" {
  name = "policy-external-secrets-${var.cluster_name}-${var.region}"
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
        Resource = "${module.jwt_secret.secret_arn}"
      }
    ]
  })
}

## Attach policy to role
resource "aws_iam_role_policy_attachment" "external_secrets" {
  policy_arn = aws_iam_policy.external_secrets.arn
  role       = aws_iam_role.pod_sa_role.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = var.cluster_name
  namespace       = var.namespace
  service_account = var.service_account
  role_arn        = aws_iam_role.pod_sa_role.arn
}
