variable "name" {
  type = string
}

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "force_delete" {
  type    = bool
  default = true
}

variable "account_id" {
  type = number
}

variable "service_role" {
  type    = string
  default = ""
}

variable "region" {
  description = "Region where the VPC will be created."
  type        = string
}
