terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.57.0"
    }  
    github = {
      source  = "integrations/github"
      version = "~> 6.2.3"
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

# Configure the GitHub Provider
provider "github" {
}
