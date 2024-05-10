####################################
#### Netwotking Creation

module "networking" {
  source                = "../modules/networking"
  project_name          = var.project_name
  region                = var.region
  environment           = var.environment
  cidr                  = "10.100.0.0/16"
  availability_zones    = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  private_subnets       = ["10.100.0.0/22", "10.100.64.0/22", "10.100.128.0/22"]
  db_subnets            = ["10.100.4.0/22", "10.100.68.0/22", "10.100.132.0/22"]
  public_subnets        = ["10.100.32.0/22", "10.100.96.0/22", "10.100.160.0/22"]
  enable_nat_gw         = false
  create_nat_gw         = false
  single_nat_gw         = false
  enable_auto_public_ip = true
}


######################################################
### KMS Key to encrypt kubernetes resources

module "kms_key_eks" {
  source              = "../modules/kms"
  region              = var.region
  environment         = var.environment
  project_name        = var.project_name
  name                = "eks"
  key_policy          = data.aws_iam_policy_document.kms_key_policy_encrypt_logs.json
  enable_key_rotation = true
}