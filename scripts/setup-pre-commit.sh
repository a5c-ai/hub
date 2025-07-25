#!/bin/bash

# Pre-commit setup script for the Hub project
# This script sets up pre-commit hooks for new developers

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

check_command() {
    if command -v "$1" &> /dev/null; then
        log "$1 is installed ✅"
        return 0
    else
        warn "$1 is not installed ❌"
        return 1
    fi
}

# Check prerequisites
log "Checking prerequisites..."

# Check Python
if ! check_command python3; then
    error "Python 3 is required. Please install Python 3 first."
    exit 1
fi

# Check pip
if ! check_command pip3; then
    error "pip3 is required. Please install pip3 first."
    exit 1
fi

# Check Go
if ! check_command go; then
    error "Go is required. Please install Go first."
    exit 1
fi

# Check Node.js and npm
if ! check_command node; then
    error "Node.js is required. Please install Node.js first."
    exit 1
fi

if ! check_command npm; then
    error "npm is required. Please install npm first."
    exit 1
fi

# Install pre-commit if not already installed
if ! check_command pre-commit; then
    log "Installing pre-commit..."
    pip3 install --user pre-commit
    if ! check_command pre-commit; then
        warn "pre-commit not found in PATH. You may need to add ~/.local/bin to your PATH"
        export PATH="$HOME/.local/bin:$PATH"
    fi
fi

# Install Go tools
log "Installing Go tools..."
go install golang.org/x/tools/cmd/goimports@latest

# Install frontend dependencies
log "Installing frontend dependencies..."
cd frontend
npm install
cd ..

# Install pre-commit hooks
log "Installing pre-commit hooks..."
pre-commit install
pre-commit install --hook-type commit-msg

# Create secrets baseline
log "Creating secrets baseline..."
if ! command -v detect-secrets &> /dev/null; then
    log "Installing detect-secrets..."
    pip3 install --user detect-secrets
fi

# Only create baseline if it doesn't exist
if [[ ! -f ".secrets.baseline" ]]; then
    detect-secrets scan . > .secrets.baseline
    log "Created secrets baseline"
else
    log "Secrets baseline already exists"
fi

# Test the setup
log "Testing pre-commit setup..."
if pre-commit run --all-files --show-diff-on-failure; then
    log "✅ Pre-commit setup completed successfully!"
else
    warn "⚠️  Pre-commit setup completed, but some hooks failed."
    warn "This is normal on first run. The hooks have automatically fixed issues."
    warn "You may need to stage the changes and commit them."
fi

log ""
log "🎉 Pre-commit hooks are now set up!"
log ""
log "Next steps:"
log "  1. Read .pre-commit-README.md for usage instructions"
log "  2. Stage any changes made by the hooks: git add -A"
log "  3. Make your first commit to test the hooks"
log ""
log "The hooks will now run automatically on every commit."
log "You can also run them manually with: pre-commit run"
