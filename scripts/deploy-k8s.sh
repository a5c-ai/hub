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
    echo "  --enable-ssh            Enable SSH Git server access (port 2222)"
    echo "  --ssh-method METHOD     SSH exposure method: 'nginx-tcp' or 'loadbalancer' (default: nginx-tcp)"
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
ENABLE_SSH=false
SSH_METHOD="nginx-tcp"

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
        --enable-ssh)
            ENABLE_SSH=true
            shift
            ;;
        --ssh-method)
            SSH_METHOD="$2"
            shift 2
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
CERT_MANAGER_AVAILABLE=$?

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

# Function to configure SSH access for Git operations
configure_ssh_access() {
    log "Configuring SSH access for Git operations..."
    
    local apply_cmd="kubectl apply"
    if [[ "$DRY_RUN" == "true" ]]; then
        apply_cmd="kubectl apply --dry-run=client"
    fi
    
    if [[ "$SSH_METHOD" == "nginx-tcp" ]]; then
        log "Configuring NGINX TCP services for SSH..."
        
        # Check if NGINX ingress controller exists
        if kubectl get ns ingress-nginx &>/dev/null; then
            if [[ -f "$CONFIG_DIR/tcp-services-configmap.yaml" ]]; then
                log "Applying TCP services ConfigMap..."
                # Replace namespace placeholder with actual namespace
                sed "s/NAMESPACE_PLACEHOLDER/$NAMESPACE/g" "$CONFIG_DIR/tcp-services-configmap.yaml" | \
                $apply_cmd -f -
                
                warn "Note: You need to update NGINX controller to use tcp-services ConfigMap"
                warn "  1. Edit deployment: kubectl edit deployment ingress-nginx-controller -n ingress-nginx"
                warn "  2. Add arg: --tcp-services-configmap=\$(POD_NAMESPACE)/tcp-services"
                warn "  3. Edit service: kubectl edit service ingress-nginx-controller -n ingress-nginx"
                warn "  4. Add port 22 to the service ports section"
            else
                error "TCP services ConfigMap not found: $CONFIG_DIR/tcp-services-configmap.yaml"
                error "Please ensure the file exists before enabling SSH with nginx-tcp method"
                exit 1
            fi
        else
            error "NGINX ingress controller namespace not found. Please install NGINX ingress first."
            exit 1
        fi
        
    elif [[ "$SSH_METHOD" == "loadbalancer" ]]; then
        log "Creating LoadBalancer service for SSH..."
        
        if [[ -f "$CONFIG_DIR/ssh-loadbalancer-service.yaml" ]]; then
            log "Applying SSH LoadBalancer service..."
            sed '/^[[:space:]]*namespace:/d' "$CONFIG_DIR/ssh-loadbalancer-service.yaml" | \
            $apply_cmd -f - -n "$NAMESPACE"
            
            log "SSH LoadBalancer service created. External IP will be assigned shortly."
            log "Check status: kubectl get service hub-ssh-service -n $NAMESPACE"
        else
            error "SSH LoadBalancer service manifest not found: $CONFIG_DIR/ssh-loadbalancer-service.yaml"
            error "Please ensure the file exists before enabling SSH with loadbalancer method"
            exit 1
        fi
    else
        error "Invalid SSH method: $SSH_METHOD. Use 'nginx-tcp' or 'loadbalancer'"
    fi
    
    log "SSH configuration completed. Git SSH will be available on port 22."
}

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
    )
    
    # Only apply certificates if cert-manager is available
    if [[ "$CERT_MANAGER_AVAILABLE" == "0" ]]; then
        manifests+=("$CONFIG_DIR/certificates.yaml")
    else
        warn "Skipping certificates.yaml - cert-manager not available"
    fi
    
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
    
    # Configure SSH if enabled
    if [[ "$ENABLE_SSH" == "true" ]]; then
        configure_ssh_access
    fi
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
    if [[ "$CERT_MANAGER_AVAILABLE" == "0" ]]; then
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
    else
        warn "Skipping certificate status check - cert-manager not available"
        warn "TLS certificates will need to be managed manually or via infrastructure"
    fi
fi

# Show SSH configuration status if enabled
if [[ "$ENABLE_SSH" == "true" ]]; then
    log ""
    log "SSH Git Server Status:"
    log "===================="
    
    if [[ "$SSH_METHOD" == "nginx-tcp" ]]; then
        log "SSH configured via NGINX TCP services"
        log "  - Port 22 will forward to backend SSH service (port 2222)"
        log "  - Remember to update NGINX controller configuration (see warnings above)"
        log ""
        log "Test SSH connectivity:"
        log "  ssh git@hub.a5c.ai"
    elif [[ "$SSH_METHOD" == "loadbalancer" ]]; then
        log "SSH configured via LoadBalancer service"
        log "  - Service: hub-ssh-service"
        log ""
        log "Check external IP status:"
        log "  kubectl get service hub-ssh-service -n $NAMESPACE"
        
        # Try to get the external IP
        EXTERNAL_IP=$(kubectl get service hub-ssh-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
        if [[ -n "$EXTERNAL_IP" ]]; then
            log "  External IP: $EXTERNAL_IP"
            log ""
            log "Test SSH connectivity:"
            log "  ssh git@$EXTERNAL_IP"
        else
            log "  External IP: <pending>"
            log "  Run the command above to check when IP is assigned"
        fi
    fi
    
    log ""
    log "Clone repositories via SSH:"
    log "  git clone git@<domain>:owner/repo.git"
fi

log "Deployment completed successfully!"
