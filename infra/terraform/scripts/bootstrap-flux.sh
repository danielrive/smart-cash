#/bin/bash

## Configure Cluster Credentials

# $1 = CLUSTER_NAME
# $2 = AWS_REGION
# $3 = GH_USER_NAME
# $4 = FLUX_REPO_NAME

echo "---------->  get eks credentials"
aws eks update-kubeconfig --name $1  --region $2

## validate if flux is installed

flux_installed=$(kubectl api-resources | grep flux)
if [ -z "$flux_installed" ]; then
  echo "---------->  flux is not installed"

  ### install flux

  echo "---------->  installing flux cli"

  curl -s https://fluxcd.io/install.sh | sudo bash

  echo "---------->  run flux bootstrap"
  flux bootstrap github \
    --owner=$3 \
    --repository=$4 \
    --path="clusters/$1/bootstrap" \
    --branch=main \
    --personal
else
  echo "---------->  flux is installed"
fi