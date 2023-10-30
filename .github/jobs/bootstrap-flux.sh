#/bin/bash

### install flux

echo "installing flux"

curl -s https://fluxcd.io/install.sh | sudo bash

## Configure Cluster Credentials
echo "get eks credentials"
aws eks update-kubeconfig --name $CLUSTER_NAME  --region $AWS_REGION

echo "run flux bootstrap"

flux bootstrap github \
  --owner=$GH_USER_NAME \
  --repository=$FLUX_REPO_NAME \
  --path="clusters/$ENVIRONMENT" \
  --personal