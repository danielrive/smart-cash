#!/bin/bash
set -euo pipefail

# Smart-Cash Deployment Validation Script
# Usage: ./check-deployment.sh [service-name]
# Examples:
#   ./check-deployment.sh user-service    # Check specific service
#   ./check-deployment.sh                 # Check all services

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="develop"
FLUX_NAMESPACE="flux-system"

# Available services
SERVICES=(
    "user-service"
    "bank-service"
    "expenses-service"
    "payment-service"
    "payment-orchestrator-service"
    "frontend-service"
)

SERVICE_NAME="${1:-}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🚀 Smart-Cash Deployment Validation${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Function to check if kubectl is available
check_prerequisites() {
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}❌ kubectl not found. Please install kubectl.${NC}"
        exit 1
    fi

    if ! command -v flux &> /dev/null; then
        echo -e "${YELLOW}⚠️  flux CLI not found. Some checks will be skipped.${NC}"
    fi

    # Check if connected to cluster
    if ! kubectl cluster-info &> /dev/null; then
        echo -e "${RED}❌ Not connected to Kubernetes cluster.${NC}"
        exit 1
    fi
}

# Function to check Flux system health
check_flux_health() {
    echo -e "${BLUE}📦 FLUX SYSTEM HEALTH${NC}"
    echo -e "${BLUE}════════════════════${NC}"

    # Check Flux controllers
    local flux_pods=$(kubectl get pods -n $FLUX_NAMESPACE --no-headers 2>/dev/null | wc -l || echo "0")
    local flux_ready=$(kubectl get pods -n $FLUX_NAMESPACE --no-headers 2>/dev/null | grep "Running" | wc -l || echo "0")

    if [[ $flux_pods -eq $flux_ready ]]; then
        echo -e "   ${GREEN}✅ Flux controllers: $flux_ready/$flux_pods running${NC}"
    else
        echo -e "   ${RED}❌ Flux controllers: $flux_ready/$flux_pods running${NC}"
    fi

    # Check GitRepository source
    if command -v flux &> /dev/null; then
        echo ""
        echo "   Git Sources:"
        flux get sources git -n $FLUX_NAMESPACE 2>/dev/null || echo "   Unable to get git sources"
    fi

    echo ""
}

# Function to check service deployment
check_service() {
    local service=$1

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🎯 SERVICE: $service${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Check Kustomization status
    echo -e "${BLUE}📋 Kustomization Status${NC}"
    echo -e "${BLUE}═══════════════════════${NC}"

    if command -v flux &> /dev/null; then
        local kustomization_status=$(flux get kustomizations -n $FLUX_NAMESPACE 2>/dev/null | grep "${service}-service" || echo "")

        if [[ -n "$kustomization_status" ]]; then
            echo "   $kustomization_status"

            if echo "$kustomization_status" | grep -q "True"; then
                echo -e "   ${GREEN}✅ Kustomization reconciled${NC}"
            else
                echo -e "   ${RED}❌ Kustomization not reconciled${NC}"
            fi
        else
            echo -e "   ${YELLOW}⚠️  Kustomization not found for $service${NC}"
        fi
    else
        echo -e "   ${YELLOW}⚠️  Flux CLI not available, skipping kustomization check${NC}"
    fi
    echo ""

    # Check ImagePolicy
    echo -e "${BLUE}🏷️  Image Policy${NC}"
    echo -e "${BLUE}═══════════════${NC}"

    local image_policy=$(kubectl get imagepolicy -n $NAMESPACE 2>/dev/null | grep "$service" || echo "")
    if [[ -n "$image_policy" ]]; then
        echo "   $image_policy"

        # Get latest image
        local latest_image=$(kubectl get imagepolicy "img-upd-$service" -n $NAMESPACE -o jsonpath='{.status.latestImage}' 2>/dev/null || echo "N/A")
        echo -e "   ${GREEN}Latest Image: $latest_image${NC}"
    else
        echo -e "   ${YELLOW}⚠️  ImagePolicy not found${NC}"
    fi
    echo ""

    # Check Deployment
    echo -e "${BLUE}🚢 Deployment Status${NC}"
    echo -e "${BLUE}══════════════════${NC}"

    local deployment_info=$(kubectl get deployment $service -n $NAMESPACE 2>/dev/null || echo "")

    if [[ -n "$deployment_info" ]]; then
        echo "   $deployment_info"

        local desired=$(echo "$deployment_info" | tail -1 | awk '{print $2}' | cut -d'/' -f2)
        local ready=$(echo "$deployment_info" | tail -1 | awk '{print $2}' | cut -d'/' -f1)

        if [[ "$desired" == "$ready" ]]; then
            echo -e "   ${GREEN}✅ All replicas ready: $ready/$desired${NC}"
        else
            echo -e "   ${RED}❌ Not all replicas ready: $ready/$desired${NC}"
        fi
    else
        echo -e "   ${RED}❌ Deployment not found${NC}"
        return
    fi
    echo ""

    # Check Pods
    echo -e "${BLUE}🎯 Pod Status${NC}"
    echo -e "${BLUE}═════════════${NC}"

    local pods=$(kubectl get pods -n $NAMESPACE -l app=$service --no-headers 2>/dev/null)

    if [[ -n "$pods" ]]; then
        echo "$pods" | while read -r line; do
            local pod_name=$(echo "$line" | awk '{print $1}')
            local status=$(echo "$line" | awk '{print $3}')
            local restarts=$(echo "$line" | awk '{print $4}')
            local age=$(echo "$line" | awk '{print $5}')

            if [[ "$status" == "Running" && "$restarts" == "0" ]]; then
                echo -e "   ${GREEN}✅ $pod_name: $status (Age: $age, Restarts: $restarts)${NC}"
            elif [[ "$status" == "Running" ]]; then
                echo -e "   ${YELLOW}⚠️  $pod_name: $status (Age: $age, Restarts: $restarts)${NC}"
            else
                echo -e "   ${RED}❌ $pod_name: $status (Age: $age, Restarts: $restarts)${NC}"
            fi
        done

        # Get image being used
        echo ""
        echo -e "   ${BLUE}Current Image:${NC}"
        local current_image=$(kubectl get pods -n $NAMESPACE -l app=$service -o jsonpath='{.items[0].spec.containers[0].image}' 2>/dev/null || echo "N/A")
        echo "   $current_image"
    else
        echo -e "   ${RED}❌ No pods found for $service${NC}"
    fi
    echo ""

    # Check for recent events/errors
    echo -e "${BLUE}🔍 Recent Events${NC}"
    echo -e "${BLUE}═══════════════${NC}"

    local recent_events=$(kubectl get events -n $NAMESPACE --field-selector involvedObject.kind=Pod --sort-by='.lastTimestamp' 2>/dev/null | grep "$service" | tail -5 || echo "")

    if [[ -n "$recent_events" ]]; then
        echo "$recent_events"
    else
        echo -e "   ${GREEN}✅ No recent events${NC}"
    fi
    echo ""
}

# Main logic
check_prerequisites

# Check Flux health first
check_flux_health

# Check service(s)
if [[ -z "$SERVICE_NAME" ]]; then
    # Check all services
    echo -e "${YELLOW}Checking all services...${NC}"
    echo ""

    for service in "${SERVICES[@]}"; do
        check_service "$service"
    done
else
    # Check specific service
    check_service "$SERVICE_NAME"
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Validation complete!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
