## Getting aws account ID 
data "aws_caller_identity" "id_account" {}

data "aws_availability_zones" "available" {
  state = "available"
}

### Github data sources

data "github_repository" "gh_gitops" {
  full_name = "danielrive/smart-cash-gitops-argocd"
}
