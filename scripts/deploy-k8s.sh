#!/bin/bash

set -e

# Configuration
ENVIRONMENT=${1:-development}
NAMESPACE="hub-${ENVIRONMENT}"
CONFIG_DIR="k8s"
HELM_RELEASE_NAME="hub"
# Use registry and version from environment variables
VERSION=${VERSION:-"latest"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Print usage
usage() {
    echo "Usage: $0 [ENVIRONMENT] [OPTIONS]"
    echo ""
    echo "Arguments:" 
    echo "  ENVIRONMENT              Environment to deploy to (default: development)"
    echo ""
    echo "Options:"
    echo "  --helm                   Use Helm for deployment instead of kubectl"
    echo "  --values FILE            Helm values file (only with --helm)"
    echo "  --dry-run               Perform a dry run"
    echo "  --skip-dependencies     Skip dependency deployments (PostgreSQL, Redis)"
    echo "  --wait                  Wait for deployments to be ready"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 development          # Deploy to development environment"
    echo "  $0 production --helm --values prod-values.yaml"
    echo "  $0 staging --dry-run    # Dry run for staging"
}

# Default options
USE_HELM=false
VALUES_FILE=""
DRY_RUN=false
SKIP_DEPENDENCIES=false
WAIT_FOR_READY=false

# Parse command line arguments
shift # Remove the environment argument
while [[ $# -gt 0 ]]; do
    case $1 in
        --helm)
            USE_HELM=true
            shift
            ;;
        --values)
            VALUES_FILE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-dependencies)
            SKIP_DEPENDENCIES=true
            shift
            ;;
        --wait)
            WAIT_FOR_READY=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

log "Deploying Hub to $ENVIRONMENT environment..."
log "Namespace: $NAMESPACE"

# Validate kubectl connection
log "Checking Kubernetes connection..."
if ! kubectl cluster-info >/dev/null 2>&1; then
    error "Cannot connect to Kubernetes cluster"
    exit 1
fi

debug "Connected to cluster: $(kubectl config current-context)"

# Check if cert-manager is installed
check_cert_manager() {
    log "Checking cert-manager installation..."
    if ! kubectl get namespace cert-manager >/dev/null 2>&1; then
        warn "cert-manager namespace not found"
        warn "To enable TLS certificates, install cert-manager first:"
        warn "  kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.9.1/cert-manager.crds.yaml"
        warn "  helm repo add jetstack https://charts.jetstack.io"
        warn "  helm repo update"
        warn "  helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.9.1"
        return 1
    fi
    
    if ! kubectl get pods -n cert-manager -l app=cert-manager --field-selector=status.phase=Running >/dev/null 2>&1; then
        warn "cert-manager pods are not running"
        warn "Check cert-manager status: kubectl get pods -n cert-manager"
        return 1
    fi
    
    log "cert-manager is installed and running"
    return 0
}

check_cert_manager

## Create or update namespace
log "Creating/updating namespace: $NAMESPACE"
if [[ "$DRY_RUN" == "true" ]]; then
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml
else
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
fi

# Configure image pull secret for private registry if credentials provided
if [[ -n "$REGISTRY" && -n "$AZURE_APPLICATION_CLIENT_ID" && -n "$AZURE_APPLICATION_CLIENT_SECRET" ]]; then
    log "Applying image pull secret for registry $REGISTRY"
    if [[ "$DRY_RUN" == "true" ]]; then
        kubectl create secret docker-registry acr-auth \
          --docker-server="$REGISTRY" \
          --docker-username="$AZURE_APPLICATION_CLIENT_ID" \
          --docker-password="$AZURE_APPLICATION_CLIENT_SECRET" \
          -n "$NAMESPACE" --dry-run=client -o yaml
    else
        kubectl create secret docker-registry acr-auth \
          --docker-server="$REGISTRY" \
          --docker-username="$AZURE_APPLICATION_CLIENT_ID" \
          --docker-password="$AZURE_APPLICATION_CLIENT_SECRET" \
          -n "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    fi
fi

# Function to apply kubectl manifests
apply_kubectl_manifests() {
    local apply_cmd="kubectl apply -f"
    if [[ "$DRY_RUN" == "true" ]]; then
        apply_cmd="kubectl apply --dry-run=client -f"
    fi
    
    log "Applying Kubernetes manifests..."
    
    # Apply in specific order for dependencies
    manifests=(
        "$CONFIG_DIR/configmap.yaml"
        "$CONFIG_DIR/secrets.yaml"
        "$CONFIG_DIR/storage.yaml"
        "$CONFIG_DIR/cert-manager-issuers.yaml"
    )
    
    if [[ "$SKIP_DEPENDENCIES" == "false" ]]; then
        manifests+=(
            "$CONFIG_DIR/postgresql-deployment.yaml"
            "$CONFIG_DIR/redis-deployment.yaml"
        )
    fi
    
    manifests+=(
        "$CONFIG_DIR/backend-deployment.yaml"
        "$CONFIG_DIR/frontend-deployment.yaml"
        "$CONFIG_DIR/services.yaml"
        "$CONFIG_DIR/ingress.yaml"
        "$CONFIG_DIR/hpa.yaml"
        "$CONFIG_DIR/network-policy.yaml"
    )
    
    for manifest in "${manifests[@]}"; do
        if [[ -f "$manifest" ]]; then
            log "Applying $manifest..."
        # Apply image substitutions for application images only; skip database and redis manifests
        if [[ -n "$REGISTRY" && "$manifest" != *postgresql* && "$manifest" != *redis* ]]; then
            sed '/^[[:space:]]*namespace:/d' "$manifest" | \
            sed -E "s#^([[:space:]]*image:[[:space:]]+)([^[:space:]]+):.*#\\1${REGISTRY}/\\2:${VERSION}#" | \
            $apply_cmd - -n "$NAMESPACE"
        else
            sed '/^[[:space:]]*namespace:/d' "$manifest" | \
            $apply_cmd - -n "$NAMESPACE"
        fi
        else
            warn "Manifest not found: $manifest"
        fi
    done
}

# Function to deploy using Helm
deploy_helm() {
    log "Deploying using Helm..."
    
    local helm_cmd="helm upgrade --install $HELM_RELEASE_NAME helm/hub"
    helm_cmd="$helm_cmd --namespace $NAMESPACE --create-namespace"
    
    if [[ -n "$VALUES_FILE" && -f "$VALUES_FILE" ]]; then
        helm_cmd="$helm_cmd --values $VALUES_FILE"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd="$helm_cmd --dry-run"
    fi
    
    if [[ "$WAIT_FOR_READY" == "true" ]]; then
        # Use DEPLOY_TIMEOUT for Helm wait timeout or default to 10m
        helm_timeout=${DEPLOY_TIMEOUT:-10m}
        helm_cmd="$helm_cmd --wait --timeout=$helm_timeout"
    fi
    
    debug "Executing: $helm_cmd"
    eval "$helm_cmd"
}

# Deploy based on chosen method
if [[ "$USE_HELM" == "true" ]]; then
    # Check if Helm is installed
    if ! command -v helm &> /dev/null; then
        error "Helm is not installed"
        exit 1
    fi
    
    deploy_helm
else
    apply_kubectl_manifests
fi

# Wait for deployments if requested and not using Helm (which has its own wait)
if [[ "$WAIT_FOR_READY" == "true" && "$USE_HELM" == "false" && "$DRY_RUN" == "false" ]]; then
    log "Waiting for deployments to be ready..."

    # Determine rollout timeout (fallback to 300s if not set)
    TIMEOUT=${DEPLOY_TIMEOUT:-300s}
    log "Using rollout timeout: $TIMEOUT"

    deployments=("hub-backend" "hub-frontend")
    if [[ "$SKIP_DEPENDENCIES" == "false" ]]; then
        deployments=("postgresql" "redis" "${deployments[@]}")
    fi

    for deployment in "${deployments[@]}"; do
        log "Waiting for deployment: $deployment"
        # Wait for deployment readiness with configurable timeout
        if ! kubectl rollout status deployment/"$deployment" -n "$NAMESPACE" --timeout="$TIMEOUT"; then
            error "Deployment of $deployment failed or timed out."
            log "Fetching pods in namespace $NAMESPACE..."
            kubectl get pods -n "$NAMESPACE"
            log "Fetching logs for $deployment pods..."
            for pod in $(kubectl get pods -n "$NAMESPACE" -l app="$deployment" -o name); do
                kubectl logs "$pod" -n "$NAMESPACE" || true
            done
            exit 1
        fi
    done
fi

# Display deployment status
if [[ "$DRY_RUN" == "false" ]]; then
    log "Deployment status:"
    kubectl get pods,svc,ingress -n "$NAMESPACE"
    
    # Show useful information
    log "Useful commands:"
    echo "  View pods: kubectl get pods -n $NAMESPACE"
    echo "  View logs: kubectl logs -n $NAMESPACE -l app=hub-backend"
    echo "  Port forward: kubectl port-forward -n $NAMESPACE svc/hub-frontend-service 3000:3000"
    
    # Try to get ingress URL
    if kubectl get ingress hub-ingress -n "$NAMESPACE" >/dev/null 2>&1; then
        INGRESS_HOST=$(kubectl get ingress hub-ingress -n "$NAMESPACE" -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "")
        if [[ -n "$INGRESS_HOST" ]]; then
            log "Application should be available at: https://$INGRESS_HOST"
        fi
    fi
    
    # Check certificate status if cert-manager is available
    if kubectl get namespace cert-manager >/dev/null 2>&1; then
        log "Certificate status:"
        if kubectl get certificate -n "$NAMESPACE" >/dev/null 2>&1; then
            kubectl get certificate -n "$NAMESPACE"
            
            # Check if TLS secret exists
            if kubectl get secret hub-azure-ssl-certificate -n "$NAMESPACE" >/dev/null 2>&1; then
                log "✓ TLS certificate secret found"
            else
                warn "✗ TLS certificate secret not found yet"
                warn "  Monitor certificate: kubectl describe certificate -n $NAMESPACE"
                warn "  Check cert-manager logs: kubectl logs -n cert-manager -l app=cert-manager"
            fi
        else
            warn "No certificates found in namespace $NAMESPACE"
        fi
    fi
fi

log "Deployment completed successfully!"
