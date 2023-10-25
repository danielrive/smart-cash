
variable "region" {
  description = "AWS region"
  type        = string
}

variable "name" {
  description = "Environment"
  type        = string
}

variable "environment" {
  description = "Environment"
  type        = string
}

variable "project_name" {
  description = "Name of kms"
  type        = string
}

variable "enable_key_rotation" {
  description = "to enable key rotation, double check the documentation, the rotation process doesnt decrypt and encrypt with the new key, rotation period is every year"
  type        = bool
  default     = true
}


variable "key_policy" {
  description = "Policy to associate with the KMS Key"
  type        = any
  default     = null
}

variable "deletion_days" {
  description = "Days to remove completely the key"
  type        = number
  default     = 30
}


