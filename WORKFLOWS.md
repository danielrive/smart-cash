# GitHub Workflows Guide

This document describes how to run the GitHub workflows in this repository.

## Prerequisites

- [GitHub CLI (gh)](https://cli.github.com/) installed
- Authenticated with GitHub: `gh auth login`
- Appropriate repository permissions

## Available Workflows

### 1. Infrastructure Deployment (Develop)

**Workflow:** `infra-deployment-develop.yaml`  
**Purpose:** Deploy infrastructure stages to the develop environment

**Trigger Methods:**

**Option A: Using the helper script**
```bash
./run-workflow.sh infra-develop --stage 1-base-stage
```

**Option B: Using gh CLI directly**
```bash
gh workflow run infra-deployment-develop.yaml -f stage_to_run="1-base-stage"
```

**Option C: Via GitHub UI**
1. Go to Actions tab
2. Select "Develop Deploy Infra"
3. Click "Run workflow"
4. Select branch and stage

**Available Stages:**
- `1-base-stage` - Base infrastructure (VPC, networking, IAM, KMS)
- `2-eks-cluster-stage` - EKS cluster and core Kubernetes resources
- `3-k8-common-stage` - Common Kubernetes and AWS resources
- `user-service` - User service infrastructure
- `bank-service` - Bank service infrastructure
- `expenses-service` - Expenses service infrastructure
- `frontend-service` - Frontend service infrastructure
- `hydradb-service` - HydraDB service infrastructure
- `payment-service` - Payment service infrastructure

**Auto-trigger:** Pushes to `develop`, `feat/*`, or `feature/*` branches that modify files in `infra/**`

---

### 2. Microservice Deployment (Develop)

**Workflow:** `service-build-push-develop.yaml`  
**Purpose:** Build and deploy microservices to the develop environment

**Trigger Methods:**

**Option A: Using the helper script**
```bash
./run-workflow.sh service-develop --service user-service
```

**Option B: Using gh CLI directly**
```bash
gh workflow run service-build-push-develop.yaml -f service_to_deploy="user-service"
```

**Option C: Via GitHub UI**
1. Go to Actions tab
2. Select "DEVELOP Microservice deploy"
3. Click "Run workflow"
4. Select service

**Available Services:**
- `user-service`
- `expenses-service`
- `bank-service`
- `frontend-service`
- `payment-service`
- `hydradb-service`

**Auto-trigger:** Pushes to `develop`, `feat/*`, or `feature/*` branches that modify files in `app/**`

**What it does:**
1. Builds Docker image for the service
2. Runs security scanning with Trivy
3. Uploads security report to GitHub
4. Pushes image to AWS ECR
5. Updates GitOps repository for FluxCD deployment

---

### 3. Microservice Deployment (Production)

**Workflow:** `service-build-push-prod.yaml`  
**Purpose:** Build and deploy microservices to the production environment

**⚠️ WARNING:** This deploys to PRODUCTION!

**Trigger Methods:**

**Option A: Using the helper script**
```bash
./run-workflow.sh service-prod --service user-service
```

**Option B: Using gh CLI directly**
```bash
gh workflow run service-build-push-prod.yaml -f service_to_deploy="user-service"
```

**Option C: Via GitHub UI**
1. Go to Actions tab
2. Select "PROD Microservice deploy"
3. Click "Run workflow"
4. Select service

**Available Services:** Same as develop environment

**Auto-trigger:** Pushes to `main2` branch that modify files in `app/**`

---

### 4. Infrastructure Destruction

**Workflow:** `infra-destroy.yaml`  
**Purpose:** Destroy infrastructure stages

**⚠️ DANGER:** This will permanently destroy infrastructure!

**Trigger Methods:**

**Option A: Using the helper script**
```bash
./run-workflow.sh infra-destroy --stage user-service
```

**Option B: Using gh CLI directly**
```bash
gh workflow run infra-destroy.yaml -f stage="user-service"
```

**Option C: Via GitHub UI**
1. Go to Actions tab
2. Select "DESTROY INFRASTRUCTURE"
3. Click "Run workflow"
4. Select stage

**Available Stages (destroy in this order):**
1. Service stages first:
   - `hydradb-service`
   - `user-service`
   - `frontend-service`
   - `expenses-service`
   - `payment-service`
   - `bank-service`
2. Then infrastructure stages:
   - `3-k8-common-stage`
   - `2-eks-cluster-stage`
   - `1-base-stage`

**Important:** Always destroy in reverse order to avoid dependency issues!

---

## Quick Reference Commands

### View workflow runs
```bash
gh run list
```

### Watch the latest workflow run
```bash
gh run watch
```

### View specific workflow runs
```bash
gh run list --workflow=infra-deployment-develop.yaml
gh run list --workflow=service-build-push-develop.yaml
```

### View logs of a specific run
```bash
gh run view <run-id> --log
```

### List all workflows
```bash
gh workflow list
```

---

## Workflow Templates

This repository uses reusable workflow templates:

### `template-run-terraform.yaml`
Template for running Terraform operations. Used by infrastructure deployment workflows.

**Parameters:**
- `AWS_REGION` - AWS region for deployment
- `ENVIRONMENT` - Environment name (develop/prod)
- `PROJECT_NAME` - Project name for tagging
- `TERRAFORM_VERSION` - Terraform version to use
- `STAGE` - Stage name to deploy
- `AWS_IAM_ROLE_GH` - IAM role for GitHub Actions

### `template-service-deploy.yaml`
Template for building and deploying microservices. Used by service deployment workflows.

**Parameters:**
- `AWS_REGION` - AWS region for deployment
- `ENVIRONMENT` - Environment name (develop/prod)
- `PROJECT_NAME` - Project name
- `SERVICE_NAME` - Service to deploy
- `TRIVY_VERSION` - Trivy version for security scanning

---

## Troubleshooting

### "gh: command not found"
Install GitHub CLI: https://cli.github.com/

### "authentication required"
Run: `gh auth login` and follow the prompts

### "workflow not found"
Make sure you're in the repository root directory

### Workflow fails immediately
Check that required secrets are configured:
- `AWS_ACCOUNT_NUMBER_DEVELOP`
- `AWS_ACCOUNT_NUMBER_PROD` (if deploying to prod)
- `GH_TOKEN_FLUXCD`

### How to check workflow status
```bash
gh run list --limit 5
```

### How to cancel a running workflow
```bash
gh run cancel <run-id>
```

---

## CI/CD Pipeline Flow

### Infrastructure Deployment
```
Push to develop → Detect changed folders → Run Terraform → Deploy infrastructure
```

### Service Deployment
```
Push to develop → Detect changed services → Build Docker image → 
Security scan → Push to ECR → Update GitOps repo → FluxCD deploys
```

### Security Scanning
All service deployments include:
- Trivy vulnerability scanning
- SARIF report upload to GitHub Security
- Pipeline fails on HIGH/CRITICAL vulnerabilities

---

## Best Practices

1. **Infrastructure deployment order:**
   - Deploy in order: 1-base → 2-eks-cluster → 3-k8-common → services

2. **Infrastructure destruction order:**
   - Destroy in reverse: services → 3-k8-common → 2-eks-cluster → 1-base

3. **Service deployment:**
   - Always deploy to develop first
   - Test thoroughly before promoting to production
   - Monitor logs after deployment

4. **Manual triggers:**
   - Use workflow_dispatch for selective deployments
   - Useful for deploying specific services without code changes
   - Good for infrastructure updates

5. **Security:**
   - Review Trivy reports in GitHub Security tab
   - Address HIGH/CRITICAL vulnerabilities before merging
   - Keep base images updated

---

## Architecture

For detailed architecture information, see [README.md](README.md):
- AWS Architecture diagram
- Kubernetes resources
- Infrastructure stages
- CI/CD flow

---

## Need Help?

- Check workflow logs: `gh run view <run-id> --log`
- View recent runs: `gh run list`
- See workflow definitions in `.github/workflows/`
- Check infrastructure code in `infra/`
- Check application code in `app/`
