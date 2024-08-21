#!/bin/bash

#echo "---> installing helm"
#curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
#chmod 700 get_helm.sh
#./get_helm.sh

echo "Checking if  argoCD already was installed"
helm status argocd -n argocd 
helm status argocd -n argocd 2>/dev/null
helm status argocd -n argocd 2>/dev/null | grep STATUS
argoInstalled=$(helm status argocd -n argocd 2>/dev/null | grep STATUS)
echo $argoInstalled
if [[ -z "$argoInstalled" ]]; then
    echo "---> argoCD no Installed"
    echo "---> getting kubeconfig"
    aws eks update-kubeconfig --name $1 --region $2
    echo "---> installing argo"
    helm repo add argo https://argoproj.github.io/argo-helm
    helm repo update
    helm install argocd argo/argo-cd --namespace argocd --create-namespace  -f ./k8-manifests/helm-argo-installation/argocd.yaml
else
    echo "---> argoCD already installed"
fi
