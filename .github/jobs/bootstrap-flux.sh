#/bin/bash

## Configure Cluster Credentials
echo "get eks credentials"
aws eks update-kubeconfig --name $CLUSTER_NAME  --region $AWS_REGION

## validate if flux is installed

flux_installed=$(kubectl api-resources | grep flux)
echo $flux_installed
if [ -z "$flux_installed" ]; then
  echo "flux is not installed"

  ### install flux

  echo "installing flux cli"

  curl -s https://fluxcd.io/install.sh | sudo bash

  echo "run flux bootstrap"
  flux bootstrap github \
    --owner=$GH_USER_NAME \
    --repository=$FLUX_REPO_NAME \
    --path="clusters/$CLUSTER_NAME" \
    --branch=main \
    --personal
else
  echo "flux is installed"
fi