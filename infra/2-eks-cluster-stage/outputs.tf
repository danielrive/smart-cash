output "cluster_endpoint" {
  value = module.eks_cluster.cluster_endpoint
}

output "cluster_oidc" {
  value = module.eks_cluster.cluster_oidc
}

output "cluster_name" {
  description = "Cluster name"
  value       = module.eks_cluster.cluster_name
}

output "cluster_security_group_id" {
  description = "Cluster security group id"
  value       = module.eks_cluster.cluster_security_group_id
}

output "created_at" {
  description = "Cluster created_at"
  value       = module.eks_cluster.created_at
}

