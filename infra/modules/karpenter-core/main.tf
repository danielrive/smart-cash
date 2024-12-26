resource "aws_cloudformation_stack" "network" {
  name = "karpenter-core-${var.cluster_name}"

  parameters = {
    ClusterName = var.cluster_name
  }
  template_body = file("./cloudformation.yaml")
  capabilities = ["CAPABILITY_NAMED_IAM"]

}