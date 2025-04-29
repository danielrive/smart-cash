#/bin/bash

## Configure Cluster Credentials

# $1 = CLUSTER_NAME
# $2 = AWS_REGION

echo "---------->  get eks credentials"
aws eks update-kubeconfig --name $1  --region $2

## validate if gateway is installed

if kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null; then
  echo "Kuberentes gateway NOT intalled"
  kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml
else 
  echo "Kuberentes gateway ALREADY intalled"
fi