#/bin/bash

## Configure Cluster Credentials

# $1 = CLUSTER_NAME
# $2 = AWS_REGION
# $3 = GH_USER_NAME
# $4 = FLUX_REPO_NAME
# $5 = Environment

echo "---------->  get eks credentials"
aws eks update-kubeconfig --name $1  --region $2

if [ $5 -eq 'production'  ]; then
  BRANCH="main"
else 
  BRANCH=$5
fi
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
    --branch=$BRANCH \
    --components-extra=image-reflector-controller,image-automation-controller \
    --personal
else
  echo "---------->  flux is installed"
fi