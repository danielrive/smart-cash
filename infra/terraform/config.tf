terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.41.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.1.0"
      
    }
  }
}

provider "flux" {
  kubernetes = {
    config_path = "~/.kube/config"
  }
  git = {
    url = "https://github.com/danielrive/smart-cash-gitops-flux"
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
