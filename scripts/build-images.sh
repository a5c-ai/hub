#!/bin/bash

set -e

# Configuration
REGISTRY=${REGISTRY:-""}
VERSION=${VERSION:-"latest"}
BUILD_CONTEXT=${BUILD_CONTEXT:-"."}
BUILD_RUNNER=${BUILD_RUNNER:-"true"}
RUNNER_ONLY=${RUNNER_ONLY:-"false"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -r, --registry REGISTRY    Container registry (e.g., myregistry.azurecr.io)"
    echo "  -v, --version VERSION      Image version tag (default: latest)"
    echo "  --runner-only              Build only the GitHub runner image"
    echo "  --skip-runner              Skip building the GitHub runner image"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  REGISTRY                   Container registry"
    echo "  VERSION                    Image version tag"
    echo "  BUILD_RUNNER               Build runner image (default: true)"
    echo ""
    echo "Examples:"
    echo "  $0                        # Build all images with default settings"
    echo "  $0 --runner-only          # Build only the runner image"
    echo "  $0 --skip-runner          # Build app images, skip runner"
    echo "  $0 -r myregistry.azurecr.io -v v1.0.0"
    echo "  REGISTRY=myregistry.azurecr.io VERSION=v1.0.0 $0"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        --runner-only)
            RUNNER_ONLY="true"
            BUILD_RUNNER="true"
            shift
            ;;
        --skip-runner)
            BUILD_RUNNER="false"
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

# Validate inputs and login to Azure ACR if needed
if [[ -z "$REGISTRY" ]]; then
    warn "No registry specified. Images will be built locally only."
    PUSH_IMAGES=false
else
    log "Using registry: $REGISTRY"
    PUSH_IMAGES=true
    # If using Azure Container Registry, perform Azure CLI and ACR login
    if [[ "$REGISTRY" == *".azurecr.io" ]]; then
        if command -v az >/dev/null 2>&1; then
            log "Logging into Azure CLI..."
            az login --service-principal -u "$AZURE_APPLICATION_CLIENT_ID" -p "$AZURE_APPLICATION_CLIENT_SECRET" --tenant "$AZURE_TENANT_ID"
            log "Logging into Azure Container Registry..."
            ACR_NAME=${REGISTRY%%.*}
            az acr login --name "$ACR_NAME"
        else
            warn "Azure CLI not found; skipping ACR login"
        fi
    fi
fi

log "Building Docker images..."
log "Version: $VERSION"

# Check if we should only build runner
if [[ "$RUNNER_ONLY" == "true" ]]; then
    log "Building only GitHub runner image..."
else
    log "Building Hub application images..."
fi

# Build backend image (skip if runner-only)
if [[ "$RUNNER_ONLY" != "true" ]]; then
log "Building backend image..."
BACKEND_IMAGE="hub/backend"
BACKEND_FULL_IMAGE="$BACKEND_IMAGE:$VERSION"

if [[ "$PUSH_IMAGES" == "true" ]]; then
    BACKEND_REGISTRY_IMAGE="$REGISTRY/$BACKEND_FULL_IMAGE"
fi

docker build -t "$BACKEND_FULL_IMAGE" \
    -f Dockerfile \
    --build-arg VERSION="$VERSION" \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --build-arg VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    "$BUILD_CONTEXT"

if [[ $? -ne 0 ]]; then
    error "Failed to build backend image"
    exit 1
fi

log "Backend image built successfully: $BACKEND_FULL_IMAGE"

# Build frontend image
log "Building frontend image..."
FRONTEND_IMAGE="hub/frontend"
FRONTEND_FULL_IMAGE="$FRONTEND_IMAGE:$VERSION"

if [[ "$PUSH_IMAGES" == "true" ]]; then
    FRONTEND_REGISTRY_IMAGE="$REGISTRY/$FRONTEND_FULL_IMAGE"
fi

docker build -t "$FRONTEND_FULL_IMAGE" \
    -f frontend/Dockerfile \
    --build-arg VERSION="$VERSION" \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --build-arg VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    frontend/

if [[ $? -ne 0 ]]; then
    error "Failed to build frontend image"
    exit 1
fi

log "Frontend image built successfully: $FRONTEND_FULL_IMAGE"

fi  # End of application images building

# Build GitHub runner image
if [[ "$BUILD_RUNNER" == "true" ]]; then
    log "Building GitHub runner image..."
    RUNNER_IMAGE="hub/github-runner"
    RUNNER_FULL_IMAGE="$RUNNER_IMAGE:$VERSION"
    
    if [[ "$PUSH_IMAGES" == "true" ]]; then
        RUNNER_REGISTRY_IMAGE="$REGISTRY/$RUNNER_FULL_IMAGE"
    fi
    
    docker build -t "$RUNNER_FULL_IMAGE" \
        -f runners/Dockerfile \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
        --build-arg VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
        .
    
    if [[ $? -ne 0 ]]; then
        error "Failed to build runner image"
        exit 1
    fi
    
    log "Runner image built successfully: $RUNNER_FULL_IMAGE"
else
    log "Skipping GitHub runner image build"
fi

# Tag and push images if registry is specified
if [[ "$PUSH_IMAGES" == "true" ]]; then
    log "Tagging images for registry..."
    
    # Tag backend image (if built)
    if [[ "$RUNNER_ONLY" != "true" ]]; then
        docker tag "$BACKEND_FULL_IMAGE" "$BACKEND_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to tag backend image"
            exit 1
        fi
        
        # Tag frontend image
        docker tag "$FRONTEND_FULL_IMAGE" "$FRONTEND_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to tag frontend image"
            exit 1
        fi
    fi
    
    # Tag runner image (if built)
    if [[ "$BUILD_RUNNER" == "true" ]]; then
        docker tag "$RUNNER_FULL_IMAGE" "$RUNNER_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to tag runner image"
            exit 1
        fi
    fi
    
    log "Pushing images to registry..."
    
    # Push backend and frontend images (if built)
    if [[ "$RUNNER_ONLY" != "true" ]]; then
        log "Pushing backend image: $BACKEND_REGISTRY_IMAGE"
        docker push "$BACKEND_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to push backend image"
            exit 1
        fi
        
        log "Pushing frontend image: $FRONTEND_REGISTRY_IMAGE"
        docker push "$FRONTEND_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to push frontend image"
            exit 1
        fi
    fi
    
    # Push runner image (if built)
    if [[ "$BUILD_RUNNER" == "true" ]]; then
        log "Pushing runner image: $RUNNER_REGISTRY_IMAGE"
        docker push "$RUNNER_REGISTRY_IMAGE"
        if [[ $? -ne 0 ]]; then
            error "Failed to push runner image"
            exit 1
        fi
    fi
    
    log "Images pushed successfully!"
    if [[ "$RUNNER_ONLY" != "true" ]]; then
        log "Backend: $BACKEND_REGISTRY_IMAGE"
        log "Frontend: $FRONTEND_REGISTRY_IMAGE"
    fi
    if [[ "$BUILD_RUNNER" == "true" ]]; then
        log "Runner: $RUNNER_REGISTRY_IMAGE"
    fi
else
    log "Images built locally:"
    if [[ "$RUNNER_ONLY" != "true" ]]; then
        log "Backend: $BACKEND_FULL_IMAGE" 
        log "Frontend: $FRONTEND_FULL_IMAGE"
    fi
    if [[ "$BUILD_RUNNER" == "true" ]]; then
        log "Runner: $RUNNER_FULL_IMAGE"
    fi
fi

# Display image information
log "Image details:"
if [[ "$RUNNER_ONLY" == "true" ]]; then
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -E "hub/github-runner" || true
else
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -E "(hub/backend|hub/frontend|hub/github-runner)" || true
fi

log "Build completed successfully!"
