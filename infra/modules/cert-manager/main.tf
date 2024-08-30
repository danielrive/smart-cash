###############################
#### IAM Role for Cert Manager

resource "aws_iam_role" "cert_manager" {
  name               = "cert-manager-${var.environment}-${var.region}"
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
          "${var.cluster_oidc}:sub" : "system:serviceaccount:security:sa-cert-manager"
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
  name   = "policy-admin-eks-${var.cluster_name}-${var.region}"
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

## Attach policy to role
resource "aws_iam_role_policy_attachment" "cert_manager" {
  policy_arn = aws_iam_policy.cert_manager.arn
  role       = aws_iam_role.cert_manager.name
}

