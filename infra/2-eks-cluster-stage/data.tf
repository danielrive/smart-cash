data "terraform_remote_state" "base" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}" 
    key    = "stage/1-base-stage/1-base-stage.tfstate"
    region  = var.region
  }
}

## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}

####################################
### Github data sources

data "github_repository" "flux-gitops" {
  full_name = "danielrive/smart-cash-gitops-flux"
}


####################################
#########  cert-manager Issuer policy

data "aws_iam_policy_document" "cert-manager-issuer" {
  statement {
    actions   = ["route53:GetChange"]
    resources = ["arn:aws:route53:::change/*"]
    effect = "Allow"
  }
  statement {
    actions   = ["route53:ChangeResourceRecordSets","route53:ListResourceRecordSets",]
    resources = ["arn:aws:route53:::hostedzone/*"]
    effect = "Allow"
  }
  statement {
    actions   = ["route53:ListHostedZonesByName",]
    resources = ["*"]
    effect = "Allow"
  }
}