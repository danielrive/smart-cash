variable "environment" {
  description = "Environment name (e.g., develop, staging, production)"
  type        = string
}

variable "secret_name" {
  description = "Name of the secret in AWS Secrets Manager"
  type        = string
}

variable "description" {
  description = "Description of the secret"
  type        = string
}

variable "project_name" {
  description = "Project name for tagging"
  type        = string
  default     = "smart-cash"
}

variable "additional_tags" {
  description = "Additional tags to apply to the secret"
  type        = map(string)
  default     = {}
}
