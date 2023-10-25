## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}

####################################
#########  KMS Policy

data "aws_iam_policy_document" "kms_key_policy_encrypt_logs" {
  statement {
    sid = "Enable IAM User Permissions"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.id_account.id}:root"]
    }
    actions   = ["kms:*", ]
    resources = ["*"]
  }

  statement {
    sid = "Enable cloudwatch Permissions"

    principals {
      type        = "Service"
      identifiers = ["logs.${var.region}.amazonaws.com"]
    }
    actions   = ["kms:Encrypt*", "kms:Decrypt*", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:Describe*"]
    resources = ["*"]
    condition {
      test     = "ArnLike"
      variable = "kms:EncryptionContext:aws:logs:arn"
      values = [
        "arn:aws:logs:${var.region}:${data.aws_caller_identity.id_account.id}:*"
      ]
    }
  }
}