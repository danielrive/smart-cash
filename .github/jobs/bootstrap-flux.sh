#/bin/bash

### install flux

echo "installing flux cli"

curl -s https://fluxcd.io/install.sh | sudo bash

## Configure Cluster Credentials
echo "get eks credentials"
aws eks update-kubeconfig --name $CLUSTER_NAME  --region $AWS_REGION

## validate if flux is installed

$fluxCRDs = kubectl api-resources

echo "run flux bootstrap"

flux bootstrap github \ # manifest for flucd configs will be stored in GitHub repo
  --owner=$GH_USER_NAME \  # Define the user name to use in GitHub
  --repository=$FLUX_REPO_NAME \ # The Github repository name where the flux manifest will be stored
  --path="clusters/$ENVIRONMENT" \ # The path where the flux manifest will be stored
  --personal  # the owner is assumed to be a GitHub user not an organization
  --branch main # Branch name 