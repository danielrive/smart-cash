data "terraform_remote_state" "base" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}" 
    key    = "stage/base/base.tfstate"
    region  = var.region
  }
}

## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}

####################################
### Github data sources

data "github_repository" "flux-gitops" {
  full_name = "danielrive/smart-cash-gitops-flux"
}
