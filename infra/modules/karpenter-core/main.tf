resource "aws_cloudformation_stack" "network" {
  name = "karpenter-core-${var.cluster_name}"

  parameters = {
    ClusterName = var.cluster_name
  }
  template_url = "https://raw.githubusercontent.com/aws/karpenter-provider-aws/${var.karpenter_version}/website/content/en/preview/getting-started/getting-started-with-karpenter/cloudformation.yaml"
  capabilities = ["CAPABILITY_NAMED_IAM"]

}