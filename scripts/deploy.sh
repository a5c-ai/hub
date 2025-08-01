#!/bin/bash

set -e

# Configuration
ENVIRONMENT=${1:-development}
DEPLOYMENT_TYPE=${DEPLOYMENT_TYPE:-kubernetes}
BUILD_IMAGES=${BUILD_IMAGES:-false}  # Skip redundant image builds in deploy step by default
RUN_TESTS=${RUN_TESTS:-false}       # Skip redundant tests in deploy step by default
REGISTRY=${REGISTRY:-""}
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "latest")}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
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

deploy_log() {
    echo -e "${MAGENTA}[DEPLOY]${NC} $1"
}

# Print usage
usage() {
    echo "Usage: $0 [ENVIRONMENT] [OPTIONS]"
    echo ""
    echo "Arguments:"
    echo "  ENVIRONMENT              Target environment (staging, production, development)"
    echo ""
    echo "Options:"
    echo "  --type TYPE              Deployment type (kubernetes, docker, terraform)"
    echo "  --no-build               Skip building Docker images"
    echo "  --no-tests               Skip running tests before deployment"
    echo "  --registry REGISTRY      Container registry for images (auto-detected based on environment if not provided)"
    echo "  --version VERSION        Version tag for deployment"
    echo "  --dry-run               Perform a dry run (preview changes)"
    echo "  --rollback              Rollback to previous version"
    echo "  --help                  Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DEPLOYMENT_TYPE          Deployment method"
    echo "  REGISTRY                 Container registry"
    echo "  VERSION                  Version tag"
    echo "  BUILD_IMAGES             Build images before deploy (true/false)"
    echo "  RUN_TESTS                Run tests before deploy (true/false)"
    echo "  DEPLOY_TIMEOUT           Timeout for rollout and Helm wait commands (e.g., 300s, 10m)"
    echo "  AZURE_APPLICATION_CLIENT_ID     Service principal client ID for Azure login"
    echo "  AZURE_APPLICATION_CLIENT_SECRET Service principal client secret for Azure login"
    echo "  AZURE_TENANT_ID                Azure AD tenant ID for Azure login"
    echo "  AZURE_RESOURCE_GROUP_NAME      Azure resource group for AKS credentials"
    echo "  AZURE_AKS_CLUSTER_NAME         AKS cluster name for kubectl context"
    echo "  AZURE_APPLICATION_CLIENT_ID     Service principal client ID for Azure login"
    echo "  AZURE_APPLICATION_CLIENT_SECRET Service principal client secret for Azure login"
    echo "  AZURE_TENANT_ID               Azure AD tenant ID for Azure login"
    echo "  AZURE_RESOURCE_GROUP_NAME     Azure resource group for AKS credentials"
    echo "  AZURE_AKS_CLUSTER_NAME        AKS cluster name for kubectl context"
    echo ""
    echo "Examples:"
    echo "  $0 staging               # Deploy to staging"
    echo "  $0 production --registry myregistry.azurecr.io --version v1.2.3"
    echo "  $0 staging --dry-run     # Preview staging deployment"
    echo "  $0 production --rollback # Rollback production"
}

# Default options
DRY_RUN=false
ROLLBACK=false

# Check for help first
if [[ "$1" == "--help" ]]; then
    usage
    exit 0
fi

# Parse command line arguments (skip first argument which is environment)
if [[ $# -gt 0 ]]; then
    shift # Remove environment argument
fi

while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            DEPLOYMENT_TYPE="$2"
            shift 2
            ;;
        --no-build)
            BUILD_IMAGES=false
            shift
            ;;
        --no-tests)
            RUN_TESTS=false
            shift
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --rollback)
            ROLLBACK=true
            shift
            ;;
        --help)
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

# Validate environment
case $ENVIRONMENT in
    development|staging|production)
        deploy_log "Deploying to $ENVIRONMENT environment"
        ;;
    *)
        error "Invalid environment: $ENVIRONMENT"
        error "Valid environments: development, staging, production"
        exit 1
        ;;
esac

# Validate deployment type
case $DEPLOYMENT_TYPE in
    kubernetes|docker|terraform)
        debug "Using deployment type: $DEPLOYMENT_TYPE"
        ;;
    *)
        error "Invalid deployment type: $DEPLOYMENT_TYPE"
        error "Valid types: kubernetes, docker, terraform"
        exit 1
        ;;
esac

deploy_log "Starting Hub deployment..."
deploy_log "Environment: $ENVIRONMENT"
deploy_log "Deployment type: $DEPLOYMENT_TYPE"
deploy_log "Version: $VERSION"
deploy_log "Registry: ${REGISTRY:-"local"}"
deploy_log "Dry run: $DRY_RUN"

# Set default container registry based on Terraform naming conventions
if [[ -z "$REGISTRY" ]]; then
    case "$ENVIRONMENT" in
        development)
            # Use registry name without hyphens to match Terraform naming (acr${local.base_name}.azurecr.io)
            REGISTRY="acrhubdevelopmentwestus3.azurecr.io"
            ;;
        staging)
            REGISTRY="acrhubstagingwestus3.azurecr.io"
            ;;
        production)
            REGISTRY="acrhubproductionwestus3.azurecr.io"
            ;;
        *)
            deploy_log "Unknown environment: $ENVIRONMENT. Container registry not set."
            ;;
    esac
    
    if [[ -n "$REGISTRY" ]]; then
        deploy_log "Using container registry for $ENVIRONMENT environment: $REGISTRY"
    fi
fi
export REGISTRY VERSION

# Set default Azure resource names based on Terraform naming conventions
case "$ENVIRONMENT" in
    development)
        export AZURE_RESOURCE_GROUP_NAME=${AZURE_RESOURCE_GROUP_NAME:-"rg-hub-development-westus3"}
        export AZURE_AKS_CLUSTER_NAME=${AZURE_AKS_CLUSTER_NAME:-"aks-hub-development-westus3-v2"}
        ;;
    staging)
        export AZURE_RESOURCE_GROUP_NAME=${AZURE_RESOURCE_GROUP_NAME:-"rg-hub-staging-westus3"}
        export AZURE_AKS_CLUSTER_NAME=${AZURE_AKS_CLUSTER_NAME:-"aks-hub-staging-westus3"}
        ;;
    production)
        export AZURE_RESOURCE_GROUP_NAME=${AZURE_RESOURCE_GROUP_NAME:-"rg-hub-production-westus3"}
        export AZURE_AKS_CLUSTER_NAME=${AZURE_AKS_CLUSTER_NAME:-"aks-hub-production-westus3"}
        ;;
    *)
        warn "Unknown environment: $ENVIRONMENT. Azure resource names not set."
        ;;
esac

if [[ -n "$AZURE_RESOURCE_GROUP_NAME" && -n "$AZURE_AKS_CLUSTER_NAME" ]]; then
    deploy_log "Azure resource names for $ENVIRONMENT environment:"
    deploy_log "  Resource Group: $AZURE_RESOURCE_GROUP_NAME"
    deploy_log "  AKS Cluster: $AZURE_AKS_CLUSTER_NAME"
fi

# Azure CLI login and AKS credentials for Kubernetes deployments
if command -v az >/dev/null 2>&1 && [[ "$DEPLOYMENT_TYPE" == "kubernetes" ]]; then
    if [[ -n "$AZURE_APPLICATION_CLIENT_ID" ]]; then
        deploy_log "Logging into Azure CLI..."
        az login --service-principal -u "$AZURE_APPLICATION_CLIENT_ID" -p "$AZURE_APPLICATION_CLIENT_SECRET" --tenant "$AZURE_TENANT_ID"
    fi
    if [[ -n "$AZURE_RESOURCE_GROUP_NAME" && -n "$AZURE_AKS_CLUSTER_NAME" ]]; then
        deploy_log "Fetching AKS credentials for cluster $AZURE_AKS_CLUSTER_NAME..."
        # If KUBECONFIG is set, write credentials there; otherwise merge into default config
        if [[ -n "$KUBECONFIG" ]]; then
            az aks get-credentials \
                --resource-group "$AZURE_RESOURCE_GROUP_NAME" \
                --name "$AZURE_AKS_CLUSTER_NAME" \
                --overwrite-existing \
                --file "$KUBECONFIG"
        else
            az aks get-credentials \
                --resource-group "$AZURE_RESOURCE_GROUP_NAME" \
                --name "$AZURE_AKS_CLUSTER_NAME" \
                --overwrite-existing
        fi
    else
        warn "AZURE_RESOURCE_GROUP_NAME or AZURE_AKS_CLUSTER_NAME not set; skipping AKS credential fetch"
    fi
fi

# Safety check for production
if [[ "$ENVIRONMENT" == "production" && "$DRY_RUN" == "false" ]]; then
    warn "You are about to deploy to PRODUCTION!"
    read -p "Are you sure you want to continue? (yes/no): " confirmation
    if [[ "$confirmation" != "yes" ]]; then
        log "Deployment cancelled."
        exit 0
    fi
fi

# Pre-deployment checks
deploy_log "Running pre-deployment checks..."

# Check if git working directory is clean (for production)
if [[ "$ENVIRONMENT" == "production" && -n "$(git status --porcelain 2>/dev/null)" ]]; then
    error "Git working directory is not clean. Commit or stash changes before production deployment."
    exit 1
fi

# Run tests if requested
if [[ "$RUN_TESTS" == "true" && "$ROLLBACK" == "false" ]]; then
    deploy_log "Running tests before deployment..."
    if ! ./scripts/test.sh --no-e2e; then
        error "Tests failed. Deployment aborted."
        exit 1
    fi
    deploy_log "Tests passed ✅"
fi

# Build images if requested
if [[ "$BUILD_IMAGES" == "true" && "$ROLLBACK" == "false" ]]; then
    deploy_log "Building Docker images..."
    
    build_args=""
    if [[ -n "$REGISTRY" ]]; then
        build_args="$build_args --registry $REGISTRY"
    fi
    build_args="$build_args --version $VERSION"
    
    if ! ./scripts/build-images.sh $build_args; then
        error "Image build failed. Deployment aborted."
        exit 1
    fi
    deploy_log "Images built successfully ✅"
fi

# Function to deploy with Kubernetes
deploy_kubernetes() {
    deploy_log "Deploying with Kubernetes..."
    
    local k8s_args="$ENVIRONMENT"
    if [[ "$DRY_RUN" == "true" ]]; then
        k8s_args="$k8s_args --dry-run"
    fi

    # Skip Kubernetes dependency deployments (PostgreSQL, Redis) when using external managed resources
    if [[ "$ENVIRONMENT" != "development" ]]; then
        k8s_args="$k8s_args --skip-dependencies"
    fi
    
    # Use Helm if available and configured
    if [[ -d "helm" ]]; then
        k8s_args="$k8s_args --helm"
        if [[ -f "helm/values-${ENVIRONMENT}.yaml" ]]; then
            k8s_args="$k8s_args --values helm/values-${ENVIRONMENT}.yaml"
        fi
    fi
    
    k8s_args="$k8s_args --wait"
    
    if [[ "$ROLLBACK" == "true" ]]; then
        deploy_log "Rolling back Kubernetes deployment..."
        if command -v helm &> /dev/null && [[ -d "helm" ]]; then
            helm rollback hub -n "hub-${ENVIRONMENT}"
        else
            kubectl rollout undo deployment/hub-backend -n "hub-${ENVIRONMENT}"
            kubectl rollout undo deployment/hub-frontend -n "hub-${ENVIRONMENT}"
        fi
    else
        ./scripts/deploy-k8s.sh $k8s_args
    fi
}

# Function to deploy with Docker
deploy_docker() {
    deploy_log "Deploying with Docker..."
    
    if [[ "$ROLLBACK" == "true" ]]; then
        error "Rollback not supported with Docker deployment"
        exit 1
    fi
    
    # Use docker-compose for local deployments
    if [[ -f "docker-compose.${ENVIRONMENT}.yml" ]]; then
        local compose_file="docker-compose.${ENVIRONMENT}.yml"
    elif [[ -f "docker-compose.yml" ]]; then
        local compose_file="docker-compose.yml"
    else
        error "No docker-compose file found"
        exit 1
    fi
    
    # Set environment variables for docker-compose
    export ENVIRONMENT="$ENVIRONMENT"
    export VERSION="$VERSION"
    export REGISTRY="$REGISTRY"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        deploy_log "Dry run - would execute:"
        echo "docker-compose -f $compose_file up -d"
    else
        docker-compose -f "$compose_file" up -d
    fi
}

# Function to deploy with Terraform
deploy_terraform() {
    deploy_log "Deploying with Terraform..."
    
    if [[ ! -d "terraform" ]]; then
        error "Terraform directory not found"
        exit 1
    fi
    
    cd terraform
    
    # Initialize Terraform
    terraform init
    
    # Select or create workspace for environment
    terraform workspace select "$ENVIRONMENT" 2>/dev/null || terraform workspace new "$ENVIRONMENT"
    
    # Set variables
    local tf_vars=""
    tf_vars="$tf_vars -var environment=$ENVIRONMENT"
    tf_vars="$tf_vars -var version=$VERSION"
    if [[ -n "$REGISTRY" ]]; then
        tf_vars="$tf_vars -var container_registry=$REGISTRY"
    fi
    
    if [[ -f "terraform.tfvars.${ENVIRONMENT}" ]]; then
        tf_vars="$tf_vars -var-file=terraform.tfvars.${ENVIRONMENT}"
    fi
    
    if [[ "$ROLLBACK" == "true" ]]; then
        warn "Terraform rollback requires manual intervention"
        log "Use 'terraform plan' and 'terraform apply' with previous state"
        exit 1
    fi
    
    # Plan deployment
    deploy_log "Planning Terraform deployment..."
    terraform plan $tf_vars -out=tfplan
    
    if [[ "$DRY_RUN" == "true" ]]; then
        deploy_log "Dry run complete. Plan saved to tfplan"
        cd ..
        return
    fi
    
    # Apply deployment
    deploy_log "Applying Terraform deployment..."
    terraform apply tfplan
    
    cd ..
}

# Deploy based on type
case $DEPLOYMENT_TYPE in
    kubernetes)
        deploy_kubernetes
        ;;
    docker)
        deploy_docker
        ;;
    terraform)
        deploy_terraform
        ;;
esac

# Post-deployment verification
if [[ "$DRY_RUN" == "false" && "$ROLLBACK" == "false" ]]; then
    deploy_log "Running post-deployment verification..."
    
    # Wait a moment for services to start
    sleep 10
    
    # Basic health check (customize based on your application)
    case $DEPLOYMENT_TYPE in
        kubernetes)
            if command -v kubectl &> /dev/null && kubectl cluster-info >/dev/null 2>&1; then
                # Check pod status
                kubectl get pods -n "hub-${ENVIRONMENT}" -l app=hub

                # Try to get service URL
                SERVICE_URL=$(kubectl get ingress hub-ingress -n "hub-${ENVIRONMENT}" -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "")
                if [[ -n "$SERVICE_URL" ]]; then
                    deploy_log "Application URL: https://$SERVICE_URL"

                    # Basic health check
                    if curl -f -s "https://$SERVICE_URL/health" >/dev/null; then
                        deploy_log "Health check passed ✅"
                    else
                        warn "Health check failed ⚠️"
                    fi
                fi
            else
                warn "kubectl cannot connect to cluster, skipping post-deployment verification"
            fi
            ;;
        docker)
            docker-compose ps
            ;;
    esac
fi

deploy_log "🎉 Deployment completed successfully!"

# Show next steps
log "Next steps:"
case $DEPLOYMENT_TYPE in
    kubernetes)
        log "  • Monitor: kubectl get pods -n hub-${ENVIRONMENT} -w"
        log "  • Logs: kubectl logs -n hub-${ENVIRONMENT} -l app=hub-backend -f"
        log "  • Rollback: $0 $ENVIRONMENT --rollback"
        ;;
    docker)
        log "  • Monitor: docker-compose logs -f"
        log "  • Stop: docker-compose down"
        ;;
    terraform)
        log "  • Monitor resources in your cloud provider console"
        log "  • Modify: Edit terraform files and re-run deployment"
        ;;
esac
