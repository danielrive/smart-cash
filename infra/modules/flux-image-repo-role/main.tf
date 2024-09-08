###############################
#### IAM Role for Cert Manager

resource "aws_iam_role" "flux_imagerepository" {
  name               = "flux-images-${var.environment}-${var.region}"
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Principal": {
        "Federated": "arn:aws:iam::${var.account_id}:oidc-provider/${var.cluster_oidc}"
      },
      "Condition": {
        "StringEquals": {
          "${var.cluster_oidc}:aud" : "sts.amazonaws.com",
          "${var.cluster_oidc}:sub" : "system:serviceaccount:flux-system:image-reflector-controller"
        }
      }
    }
  ]
}
EOF
}

###############################
#### Role Policy

resource "aws_iam_policy" "cert_manager" {
  name   = "policy-certmanager-${var.cluster_name}-${var.region}"
  path   = "/"
  policy = <<EOF
    {
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Action": "route53:GetChange",
        "Resource": "arn:aws:route53:::change/*"
        },
        {
        "Effect": "Allow",
        "Action": [
            "route53:ChangeResourceRecordSets",
            "route53:ListResourceRecordSets"
        ],
        "Resource": "arn:aws:route53:::hostedzone/*"
        },
        {
        "Effect": "Allow",
        "Action": "route53:ListHostedZonesByName",
        "Resource": "*"
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
  role       = aws_iam_role.flux_imagerepository.name
}

