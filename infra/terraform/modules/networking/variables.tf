variable "project_name" {
  type = any
}

variable "environment" {
  type = any
}

variable "cidr" {
  type = any
}

variable "availability_zones" {
  type = any
}

variable "private_subnets" {
  type = any
}

variable "public_subnets" {
  type = any
}

variable "db_subnets" {
  type = any
}

variable "create_db_subnet_group" {
  type    = any
  default = true
}

variable "create_nat_gw" {
  type    = any
  default = true
}

variable "enable_nat_gw" {
  type    = any
  default = true
}

variable "single_nat_gw" {
  type    = any
  default = true
}

variable "one_nat_per_az" {
  type    = any
  default = false
}

variable "tags" {
  type    = any
  default = {}
}

variable "region" {
  type = string
}