###################################
####### Networking outputs

output "vpc_id" {
  value = module.networking.vpc.id
}


output "public_subnets" {
  value = module.networking.main.public_subnets
}


output "private_subnets" {
  value = module.networking.main.public_subnets
}

output "kms_eks_arn" {
  value = module.kms_key_eks.kms_arn
}

output "kms_eks_id" {
  value = module.kms_key_eks.kms_id
}