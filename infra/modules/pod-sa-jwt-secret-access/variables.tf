variable "environment" {
  description = "Environment name matching smart-cash/jwt-secret/<environment>"
  type        = string
}

variable "iam_role_id" {
  description = "IAM role id for the workload service account (aws_iam_role.<name>.id)"
  type        = string
}

variable "name_suffix" {
  description = "Short identifier for the policy name (e.g. bank, user)"
  type        = string
}
