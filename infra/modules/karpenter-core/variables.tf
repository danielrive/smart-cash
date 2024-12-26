variable "cluster_name" {
  description = "eks cluster name"
  type        = string
}

variable "karpenter_version" {
  description = "karpenter version to use"
  type        = string
}

variable "environment" {
  description = "env name"
  type        = string
}

variable "account_number" {
  description = "aws account number"
  type        = number
}