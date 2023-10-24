variable "environment" {
  description = "This is the environment where all the infra will be created"
  type        = string
}
variable "region" {
  description = "Region where the VPC will be created."
  type        = string
}

variable "project_name" {
  type = string
}

variable "userRoleARN" {
  type = string
}

variable "cluster_version" {
  type = string
}

variable "kms_arn" {
  type = string
}

variable "cluster_enabled_log_types" {
  description = "A list of the desired control plane logs to enable. For more information, see Amazon EKS Control Plane Logging documentation (https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html)"
  type        = list(string)
  default     = ["audit", "api", "authenticator"]
}

variable "subnet_ids" {
  type = list(any)
}

variable "desired_nodes" {
  type    = number
  default = 1
}

variable "retention_control_plane_logs" {
  description = "number of day that cloudwatch will retain the logs for control plane"
  type        = number
  default     = 30
}


variable "instance_type_worker_nodes" {
  description = "a list with the instances types to use for eks worker nodes"
  type        = list(any)
}

variable "AMI_for_worker_nodes" {
  description = "the AWS AMI to use in the worker nodes"
  type        = string
}

variable "min_instances_node_group" {
  description = "minimum number of instance to use in the node group"
  type        = number
}

variable "max_instances_node_group" {
  description = "max number of instance to use in the node group"
  type        = number
}

variable "private_endpoint_api" {
  description = "Whether the Amazon EKS private API server endpoint is enabled"
  type        = bool
  default     = true
}

variable "public_endpoint_api" {
  description = "Whether the Amazon EKS private API server endpoint is enabled"
  type        = bool
  default     = false
}



