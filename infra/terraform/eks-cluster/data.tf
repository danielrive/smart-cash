data "terraform_remote_state" "base" {
  backend = "s3"
  config = {
    bucket = "${var.project_name}-tf-state-lock-${var.environment}-${var.region}" 
    key    = "base/base.tfstate"
    region  = var.region
  }
}