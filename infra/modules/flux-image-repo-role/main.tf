###############################
#### IAM Role for Cert Manager

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


## Policy for the role
resource "aws_iam_policy" "allow_ecr" {
  name = "ecr-flux-images-${var.environment}-${var.region}"
  path = "/"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AllowPull",
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
        ],
        Resource = "*"
      }
    ]
  })
  }
  
## attach the policy
resource "aws_iam_role_policy_attachment" "flux_imageupdate" {
  policy_arn = aws_iam_policy.allow_ecr.arn
  role       = aws_iam_role.pod_sa_role.name
}

resource "aws_eks_pod_identity_association" "association" {
  cluster_name    = var.cluster_name
  namespace       = var.namespace
  service_account = var.service_account
  role_arn        = aws_iam_role.pod_sa_role.arn
}