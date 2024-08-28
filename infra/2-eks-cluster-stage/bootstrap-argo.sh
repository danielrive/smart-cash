#!/bin/bash

# $1 = CLUSTER_NAME
# $2 = AWS_REGION
# $3 = REPO URL
# $4 = Environment
# $5 = EKS Cluster endpoint

#### Install ArgoCD CLI
echo "---> install argocli"
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
rm argocd-linux-amd64
## Get Kubeconfig to make calls to K8 API
echo "---> getting kubeconfig"
aws eks update-kubeconfig --name $1 --region $2
## Set the ns to argocd, argo cli doesnt allow to pass the ns directly
echo "---> set-up context ns to argocd"
kubectl config set-context --current --namespace=argocd 
## Add the repository, GH TOKEN is got from GH actions secrets by envs
echo "---> add github repo"
argocd repo add $3 --password $GITHUB_TOKEN --username argobot --core
## Create the first app, the will be the main application that will contain other apps
echo "---> add main app"
argocd app create main-app --core --directory-recurse --repo $3 --revision $4 --path clusters/$1/bootstrap --dest-namespace argocd --dest-server $5 --sync-policy auto --self-heal
    