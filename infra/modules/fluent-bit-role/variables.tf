variable "environment" {
  description = "This is the environment where all the infra will be created"
  type        = string
}
variable "region" {
  description = "Region where the VPC will be created."
  type        = string
}

variable "cluster_name" {
  type = string
}

variable "cluster_oidc" {
  type = string
}


variable "account_number" {
  type = number
}
