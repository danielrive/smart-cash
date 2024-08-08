data "terraform_remote_state" "eks" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}"
    key    = "stage/2-eks-cluster-stage/2-eks-cluster-stage.tfstate"
    region = var.region
  }
}

####################################
### Github data sources

data "github_repository" "flux-gitops" {
  full_name = "danielrive/smart-cash-gitops-flux"
}


## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}
