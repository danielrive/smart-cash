#!/bin/bash

## Install HELM in the server
echo "---> installing helm"
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
## Get Kubeconfig to validate the HELM releases installed
echo "---> getting kubeconfig"
aws eks update-kubeconfig --name $1 --region $2
argoInstalled=$(helm list -n argocd --filter argocd --output json | jq -r '.[0].status')

### Check if the argo Release is present in the cluster
if [[ "$argoInstalled" == "deployed" ]]; then
    echo "---> argoCD already installed"
elif [[ "$argoInstalled" == "null" ]]; then
    ## Install Argocd chart
    echo "---> argoCD no Installed"
    echo "---> installing argo"
    helm repo add argocd https://argoproj.github.io/argo-helm
    helm repo update
    helm install argocd argocd/argo-cd --namespace argocd --create-namespace  -f ./k8-manifests/helm-argo-installation/argocd-no-ha.yaml
else
    ### Throw and error, unknow status of release
    echo "Uknow status of HELM release"
    exit 11
fi
