variable "cluster_name" {
  description = "eks cluster name"
  type        = string
}

variable "karpenter_version" {
  description = "karpenter version to use"
  type        = string
}