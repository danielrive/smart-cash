output "cluster_endpoint" {
  value = aws_eks_cluster.kube_cluster.endpoint
}

output "cluster_oidc" {
  value = replace(aws_eks_cluster.kube_cluster.identity[0].oidc[0].issuer, "https://", "")
}

output "cluster_name" {
  description = "Cluster name"
  value       = aws_eks_cluster.kube_cluster.id
}

output "cluster_security_group_id" {
  description = "Cluster security group id"
  value       = aws_eks_cluster.kube_cluster.vpc_config[0].cluster_security_group_id
}