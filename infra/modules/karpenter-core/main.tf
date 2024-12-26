resource "aws_cloudformation_stack" "network" {
  name = "karpenter-core-${var.cluster_name}"

  parameters = {
    ClusterName = var.cluster_name
  }
  template_url = "./cloudformation.yaml"
  capabilities = ["CAPABILITY_NAMED_IAM"]

}