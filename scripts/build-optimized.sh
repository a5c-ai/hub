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

log "Starting optimized Hub build process..."
log "Build environment: $BUILD_ENV"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Skip dependency installation in CI if already cached
if [[ "$CI" == "true" && -f "go.sum" && -d "frontend/node_modules" ]]; then
    log "Dependencies already cached, skipping installation..."
else
    log "Installing dependencies..."
    ./scripts/install.sh
fi

# Build Go backend with optimizations
log "Building Go backend..."
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}

# Set build flags for Go with optimizations
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_DATE"
LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"

# Set build environment optimizations
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

debug "Building with flags: $LDFLAGS"
debug "Build optimizations: CGO_ENABLED=$CGO_ENABLED, GOOS=$GOOS, GOARCH=$GOARCH"

# Build the main server binary with parallel compilation
GOMAXPROCS=${GOMAXPROCS:-$(nproc)} go build \
    -ldflags "$LDFLAGS" \
    -trimpath \
    -buildvcs=false \
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
    
    # Install frontend dependencies if not already installed (CI optimization)
    if [[ ! -d "node_modules" ]]; then
        log "Installing frontend dependencies..."
        # Use npm ci for faster, reliable, reproducible builds
        npm ci --production=false --prefer-offline --no-audit
    else
        log "Frontend dependencies already installed, skipping..."
    fi
    
    # Set environment variables for build with optimizations
    export NODE_ENV=$BUILD_ENV
    export NEXT_TELEMETRY_DISABLED=1
    export NODE_OPTIONS="--max-old-space-size=4096"
    
    # Build the frontend with optimizations
    if [[ "$CI" == "true" ]]; then
        # CI-specific optimizations
        npm run build --silent
    else
        npm run build
    fi
    
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
    # file "$OUTPUT_DIR/$BINARY_NAME"
fi

# Skip validation in CI to save time
if [[ "$CI" != "true" ]]; then
    # Run a quick validation
    log "Validating build..."
    if "./$OUTPUT_DIR/$BINARY_NAME" --version >/dev/null 2>&1; then
        log "✅ Backend binary validation passed"
    else
        warn "⚠️  Backend binary validation failed (--version flag not supported)"
    fi
fi

log "🎉 Build process completed successfully!"
log "Run with: ./$OUTPUT_DIR/$BINARY_NAME"