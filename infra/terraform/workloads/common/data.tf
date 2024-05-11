data "terraform_remote_state" "base" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}" 
    key    = "base/base.tfstate"
    region  = var.region
  }
}

data "terraform_remote_state" "eks" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}" 
    key    = "stage/eks-cluster/eks-cluster.tfstate"
    region  = var.region
  }
}

## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}

