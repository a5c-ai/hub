#!/bin/bash

set -e

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

log "Installing Hub dependencies..."

# Install Go dependencies
if [[ -f "go.mod" ]]; then
    log "Installing Go dependencies..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    log "Found Go version: $GO_VERSION"
    
    # Download and verify Go modules
    go mod download
    go mod verify
    go mod tidy
    
    log "✅ Go dependencies installed successfully"
else
    warn "No go.mod found, skipping Go dependencies"
fi

# Install frontend dependencies
if [[ -d "frontend" && -f "frontend/package.json" ]]; then
    log "Installing frontend dependencies..."
    
    cd frontend
    
    # Check if Node.js is installed
    if ! command -v node &> /dev/null; then
        error "Node.js is not installed. Please install Node.js 18 or later."
        exit 1
    fi
    
    # Check if npm is installed
    if ! command -v npm &> /dev/null; then
        error "npm is not installed. Please install npm."
        exit 1
    fi
    
    # Check Node.js version
    NODE_VERSION=$(node --version)
    NPM_VERSION=$(npm --version)
    log "Found Node.js version: $NODE_VERSION"
    log "Found npm version: $NPM_VERSION"
    
    # Clean install for reproducible builds
    if [[ -f "package-lock.json" ]]; then
        npm ci
    else
        npm install
    fi
    
    log "✅ Frontend dependencies installed successfully"
    
    cd ..
else
    warn "No frontend directory or package.json found, skipping frontend dependencies"
fi

# Install Python dependencies if requirements.txt exists
if [[ -f "requirements.txt" ]]; then
    log "Installing Python dependencies..."
    
    # Check if Python is installed
    if ! command -v python3 &> /dev/null && ! command -v python &> /dev/null; then
        error "Python is not installed. Please install Python 3.8 or later."
        exit 1
    fi
    
    # Determine Python command
    PYTHON_CMD="python3"
    if ! command -v python3 &> /dev/null; then
        PYTHON_CMD="python"
    fi
    
    # Check if pip is installed
    if ! command -v pip3 &> /dev/null && ! command -v pip &> /dev/null; then
        error "pip is not installed. Please install pip."
        exit 1
    fi
    
    # Determine pip command
    PIP_CMD="pip3"
    if ! command -v pip3 &> /dev/null; then
        PIP_CMD="pip"
    fi
    
    PYTHON_VERSION=$($PYTHON_CMD --version)
    log "Found Python version: $PYTHON_VERSION"
    
    # Install Python dependencies
    $PIP_CMD install -r requirements.txt
    
    log "✅ Python dependencies installed successfully"
else
    debug "No requirements.txt found, skipping Python dependencies"
fi

# Install development tools if needed
if [[ -f ".pre-commit-config.yaml" ]] && command -v pre-commit &> /dev/null; then
    log "Installing pre-commit hooks..."
    pre-commit install
    log "✅ Pre-commit hooks installed"
fi

# Verify database tools (optional)
if command -v psql &> /dev/null; then
    PSQL_VERSION=$(psql --version | head -n1)
    log "Found PostgreSQL client: $PSQL_VERSION"
else
    warn "PostgreSQL client (psql) not found. Database operations may not work."
fi

if command -v redis-cli &> /dev/null; then
    REDIS_VERSION=$(redis-cli --version)
    log "Found Redis client: $REDIS_VERSION"
else
    warn "Redis client (redis-cli) not found. Cache operations may not work."
fi

# Check for Docker (optional for development)
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    log "Found Docker: $DOCKER_VERSION"
else
    debug "Docker not found (optional for development)"
fi

# Check for Kubernetes tools (optional for deployment)
if command -v kubectl &> /dev/null; then
    KUBECTL_VERSION=$(kubectl version --client --output=yaml 2>/dev/null | grep gitVersion | cut -d'"' -f4)
    log "Found kubectl: $KUBECTL_VERSION"
fi

if command -v helm &> /dev/null; then
    HELM_VERSION=$(helm version --short --client)
    log "Found Helm: $HELM_VERSION"
fi

log "🎉 Dependencies installation completed successfully!"

# Display summary
log "Summary:"
log "  Go modules: $(test -f go.mod && echo "✅ Installed" || echo "❌ Not found")"
log "  Frontend packages: $(test -d frontend/node_modules && echo "✅ Installed" || echo "❌ Not found")"
log "  Python packages: $(test -f requirements.txt && echo "✅ Installed" || echo "❌ Not applicable")"

log ""
log "Next steps:"
log "  • Run './scripts/dev-run.sh' to start development server"
log "  • Run './scripts/build.sh' to build the application"
log "  • Run './scripts/test.sh' to run tests"