variable "region" {
  description = "AWS region"
  type        = string
}

variable "environment" {
  description = "Environment"
  type        = string
}

variable "project_name" {
  description = "project name"
  type        = string
}

variable "gh_token" {
  type = string
  sensitive = true  # Mark the variable as sensitive
}