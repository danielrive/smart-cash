terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.41.0"
    }
  }
}

provider "aws" {
  region = var.region
  default_tags {
    tags = {
      "Environment" = var.environment
      "Region"      = var.region
    }
  }
}

