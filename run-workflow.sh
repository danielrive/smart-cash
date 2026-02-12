#!/bin/bash

# Helper script to trigger GitHub workflows
# Requires: gh CLI installed and authenticated (gh auth login)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Usage: $0 [workflow-name] [options]"
    echo ""
    echo "Available workflows:"
    echo "  infra-develop     Deploy infrastructure (develop environment)"
    echo "  service-develop   Deploy microservice (develop environment)"
    echo "  service-prod      Deploy microservice (production environment)"
    echo "  infra-destroy     Destroy infrastructure"
    echo ""
    echo "Examples:"
    echo "  $0 infra-develop --stage 1-base-stage"
    echo "  $0 service-develop --service user-service"
    echo "  $0 service-prod --service user-service"
    echo "  $0 infra-destroy --stage user-service"
    echo ""
}

WORKFLOW=$1

# Show help if requested or no arguments
if [ "$WORKFLOW" = "" ] || [ "$WORKFLOW" = "--help" ] || [ "$WORKFLOW" = "-h" ]; then
    usage
    exit 0
fi

shift

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub CLI${NC}"
    echo "Please run: gh auth login"
    exit 1
fi

case "$WORKFLOW" in
    infra-develop)
        STAGE=""
        while [[ $# -gt 0 ]]; do
            case $1 in
                --stage)
                    STAGE="$2"
                    shift 2
                    ;;
                *)
                    echo -e "${RED}Unknown option: $1${NC}"
                    exit 1
                    ;;
            esac
        done

        if [ -z "$STAGE" ]; then
            echo -e "${YELLOW}Available stages:${NC}"
            echo "  - 1-base-stage"
            echo "  - 2-eks-cluster-stage"
            echo "  - 3-k8-common-stage"
            echo "  - user-service"
            echo "  - bank-service"
            echo "  - expenses-service"
            echo "  - frontend-service"
            echo "  - hydradb-service"
            echo "  - payment-service"
            read -p "Enter stage name: " STAGE
        fi

        echo -e "${GREEN}Triggering infrastructure deployment for stage: $STAGE${NC}"
        gh workflow run infra-deployment-develop.yaml -f stage_to_run="$STAGE"
        ;;

    service-develop)
        SERVICE=""
        while [[ $# -gt 0 ]]; do
            case $1 in
                --service)
                    SERVICE="$2"
                    shift 2
                    ;;
                *)
                    echo -e "${RED}Unknown option: $1${NC}"
                    exit 1
                    ;;
            esac
        done

        if [ -z "$SERVICE" ]; then
            echo -e "${YELLOW}Available services:${NC}"
            echo "  - user-service"
            echo "  - expenses-service"
            echo "  - bank-service"
            echo "  - frontend-service"
            echo "  - payment-service"
            echo "  - hydradb-service"
            read -p "Enter service name: " SERVICE
        fi

        echo -e "${GREEN}Triggering service deployment for: $SERVICE${NC}"
        gh workflow run service-build-push-develop.yaml -f service_to_deploy="$SERVICE"
        ;;

    service-prod)
        SERVICE=""
        while [[ $# -gt 0 ]]; do
            case $1 in
                --service)
                    SERVICE="$2"
                    shift 2
                    ;;
                *)
                    echo -e "${RED}Unknown option: $1${NC}"
                    exit 1
                    ;;
            esac
        done

        if [ -z "$SERVICE" ]; then
            echo -e "${YELLOW}Available services:${NC}"
            echo "  - user-service"
            echo "  - expenses-service"
            echo "  - bank-service"
            echo "  - frontend-service"
            echo "  - payment-service"
            echo "  - hydradb-service"
            read -p "Enter service name: " SERVICE
        fi

        echo -e "${GREEN}Triggering PRODUCTION service deployment for: $SERVICE${NC}"
        echo -e "${RED}WARNING: This will deploy to PRODUCTION!${NC}"
        read -p "Are you sure? (yes/no): " CONFIRM
        if [ "$CONFIRM" != "yes" ]; then
            echo "Cancelled."
            exit 0
        fi

        gh workflow run service-build-push-prod.yaml -f service_to_deploy="$SERVICE"
        ;;

    infra-destroy)
        STAGE=""
        while [[ $# -gt 0 ]]; do
            case $1 in
                --stage)
                    STAGE="$2"
                    shift 2
                    ;;
                *)
                    echo -e "${RED}Unknown option: $1${NC}"
                    exit 1
                    ;;
            esac
        done

        if [ -z "$STAGE" ]; then
            echo -e "${YELLOW}Available stages to destroy:${NC}"
            echo "  - hydradb-service"
            echo "  - user-service"
            echo "  - frontend-service"
            echo "  - expenses-service"
            echo "  - payment-service"
            echo "  - bank-service"
            echo "  - 3-k8-common-stage"
            echo "  - 2-eks-cluster-stage"
            echo "  - 1-base-stage"
            read -p "Enter stage name to destroy: " STAGE
        fi

        echo -e "${RED}WARNING: This will DESTROY infrastructure for stage: $STAGE${NC}"
        read -p "Are you sure? Type 'destroy' to confirm: " CONFIRM
        if [ "$CONFIRM" != "destroy" ]; then
            echo "Cancelled."
            exit 0
        fi

        echo -e "${GREEN}Triggering infrastructure destruction for stage: $STAGE${NC}"
        gh workflow run infra-destroy.yaml -f stage="$STAGE"
        ;;

    *)
        echo -e "${RED}Unknown workflow: $WORKFLOW${NC}"
        usage
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Workflow triggered successfully!${NC}"
echo "View workflow runs with: gh run list"
echo "Watch latest run with: gh run watch"
