###############################
#### IAM Role for Fluent-bit

resource "aws_iam_role" "fluent_bit" {
  name = "role-fluent-bit-${var.environment}"
  path = "/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::${var.account_number}:oidc-provider/${var.cluster_oidc}"
        },
        Action = "sts:AssumeRoleWithWebIdentity",
        Condition = {
          StringEquals = {
            "${var.cluster_oidc}:aud" : "sts.amazonaws.com",
            "${var.cluster_oidc}:sub" : "system:serviceaccount:fluent-bit:fluent-bit"
          }
        }
      }
    ]
  })
}

resource "aws_iam_policy" "fluent_bit" {
  name        = "policy-fluent-bit-${var.environment}"
  path        = "/"
  description = "policy for k8 service account"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Effect   = "Allow"
        Resource = ["arn:aws:logs:${var.region}:${var.account_number}:log-group:/aws/eks/${var.cluster_name}/workloads:*"]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "att_policy_role1" {
  policy_arn = aws_iam_policy.fluent_bit.arn
  role       = aws_iam_role.fluent_bit.name
}
