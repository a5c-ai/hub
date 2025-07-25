#!/bin/bash

set -e

# Configuration
BUILD_ENV=${BUILD_ENV:-production}
OUTPUT_DIR=${OUTPUT_DIR:-./dist}
BINARY_NAME=${BINARY_NAME:-hub}

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

log "Starting Hub build process..."
log "Build environment: $BUILD_ENV"

# Install dependencies first
log "Installing dependencies..."
./scripts/install.sh

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build Go backend
log "Building Go backend..."
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}

# Set build flags for Go
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_DATE"
LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"

debug "Building with flags: $LDFLAGS"

# Build the main server binary with optimizations
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOCACHE=/tmp/go-build-cache
export GOMAXPROCS=4
export GOGC=50
export GOFLAGS="-p=4 -buildvcs=false"

timeout 25m go build \
    -ldflags "$LDFLAGS" \
    -trimpath \
    -tags netgo \
    -installsuffix netgo \
    -o "$OUTPUT_DIR/$BINARY_NAME" \
    ./cmd/server

if [[ $? -ne 0 ]]; then
    error "Failed to build Go backend"
    exit 1
fi

log "Go backend built successfully: $OUTPUT_DIR/$BINARY_NAME"

# Build frontend if it exists
if [[ -d "frontend" && -f "frontend/package.json" ]]; then
    log "Building frontend..."
    
    cd frontend
    
    # Install frontend dependencies if not already installed
    if [[ ! -d "node_modules" ]]; then
        log "Installing frontend dependencies..."
        # Use npm ci with optimizations for CI and timeout
        timeout 20m npm ci --production=false --prefer-offline --no-audit --no-fund --progress=false
    fi
    
    # Set environment variables for build
    export NODE_ENV=$BUILD_ENV
    export NEXT_TELEMETRY_DISABLED=1
    export NODE_OPTIONS="--max-old-space-size=8192"
    
    # Build the frontend with optimizations and timeout prevention
    export DISABLE_COLLECT_BUILD_TRACES=1
    export NEXT_TELEMETRY_DISABLED=1
    export NEXT_BUILD_DISABLE_STATIC_OPTIMIZATION=true
    export NEXT_PARALLEL=false
    
    # Build with timeout and better error handling
    timeout 30m npm run build
    
    if [[ $? -ne 0 ]]; then
        error "Failed to build frontend"
        exit 1
    fi
    
    log "Frontend built successfully"
    
    # Copy frontend build to output directory
    if [[ -d ".next" ]]; then
        log "Copying frontend build artifacts..."
        cp -r .next "../$OUTPUT_DIR/frontend-build"
        cp -r public "../$OUTPUT_DIR/frontend-public" 2>/dev/null || true
        cp package.json "../$OUTPUT_DIR/frontend-package.json"
    fi
    
    cd ..
else
    warn "Frontend directory not found or package.json missing, skipping frontend build"
fi

# Build information
log "Build completed successfully!"
log "Artifacts:"
log "  Backend binary: $OUTPUT_DIR/$BINARY_NAME"
if [[ -d "$OUTPUT_DIR/frontend-build" ]]; then
    log "  Frontend build: $OUTPUT_DIR/frontend-build"
fi

# Display binary information
if [[ -f "$OUTPUT_DIR/$BINARY_NAME" ]]; then
    log "Binary information:"
    ls -lh "$OUTPUT_DIR/$BINARY_NAME"
    file "$OUTPUT_DIR/$BINARY_NAME"
fi

# Run a quick validation
log "Validating build..."
if "./$OUTPUT_DIR/$BINARY_NAME" --version >/dev/null 2>&1; then
    log "✅ Backend binary validation passed"
else
    warn "⚠️  Backend binary validation failed (--version flag not supported)"
fi

log "🎉 Build process completed successfully!"
log "Run with: ./$OUTPUT_DIR/$BINARY_NAME"