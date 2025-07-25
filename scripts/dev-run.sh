#!/bin/bash

set -e

# Configuration
DEV_ENV=${DEV_ENV:-development}
BACKEND_PORT=${BACKEND_PORT:-8080}
FRONTEND_PORT=${FRONTEND_PORT:-3000}
LOG_LEVEL=${LOG_LEVEL:-4}

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

backend_log() {
    echo -e "${MAGENTA}[BACKEND]${NC} $1"
}

frontend_log() {
    echo -e "${BLUE}[FRONTEND]${NC} $1"
}

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --backend-only         Run only the backend server"
    echo "  --frontend-only        Run only the frontend server"  
    echo "  --backend-port PORT    Backend port (default: 8080)"
    echo "  --frontend-port PORT   Frontend port (default: 3000)"
    echo "  --log-level LEVEL      Log level (debug, info, warn, error)"
    echo "  --no-deps              Skip dependency installation"
    echo "  --help                 Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DEV_ENV                Development environment (default: development)"
    echo "  BACKEND_PORT           Backend server port"
    echo "  FRONTEND_PORT          Frontend server port" 
    echo "  LOG_LEVEL              Application log level"
}

# Default options
BACKEND_ONLY=false
FRONTEND_ONLY=false
SKIP_DEPS=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --backend-only)
            BACKEND_ONLY=true
            shift
            ;;
        --frontend-only)
            FRONTEND_ONLY=true
            shift
            ;;
        --backend-port)
            BACKEND_PORT="$2"
            shift 2
            ;;
        --frontend-port)
            FRONTEND_PORT="$2"
            shift 2
            ;;
        --log-level)
            LOG_LEVEL="$2"
            shift 2
            ;;
        --no-deps)
            SKIP_DEPS=true
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

# Validate conflicting options
if [[ "$BACKEND_ONLY" == "true" && "$FRONTEND_ONLY" == "true" ]]; then
    error "Cannot specify both --backend-only and --frontend-only"
    exit 1
fi

log "Starting Hub development servers..."
log "Environment: $DEV_ENV"
log "Backend port: $BACKEND_PORT"
log "Frontend port: $FRONTEND_PORT"
log "Log level: $LOG_LEVEL"

# Install dependencies if not skipped
if [[ "$SKIP_DEPS" == "false" ]]; then
    log "Installing dependencies..."
    ./scripts/install.sh
else
    debug "Skipping dependency installation"
fi

# Create config file for development if it doesn't exist
if [[ ! -f "config.yaml" && -f "config.example.yaml" ]]; then
    log "Creating development config from example..."
    cp config.example.yaml config.yaml
fi

# Setup development environment variables
export ENVIRONMENT="$DEV_ENV"
export PORT="$BACKEND_PORT"
export LOG_LEVEL="$LOG_LEVEL"
export GIN_MODE="debug"

# Database configuration for development
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_NAME="${DB_NAME:-hub}"
export DB_USER="${DB_USER:-hub}"
export DB_PASSWORD="${DB_PASSWORD:-password}"

# Function to run backend
run_backend() {
    backend_log "Starting Go backend server on port $BACKEND_PORT..."
    
    # Build and run with hot reload using air if available
    if command -v air &> /dev/null; then
        backend_log "Using air for hot reload"
        air -c .air.toml
    else
        backend_log "Hot reload not available (install 'air' for better development experience)"
        backend_log "Building and starting server..."
        go run ./cmd/server
    fi
}

# Function to run frontend  
run_frontend() {
    if [[ ! -d "frontend" ]]; then
        warn "Frontend directory not found, skipping frontend server"
        return
    fi
    
    frontend_log "Starting Next.js frontend server on port $FRONTEND_PORT..."
    
    cd frontend
    
    # Set frontend environment variables
    export NODE_ENV="development"
    export PORT="$FRONTEND_PORT"
    export NEXT_TELEMETRY_DISABLED=1
    export NEXT_PUBLIC_API_URL="http://localhost:$BACKEND_PORT/api/v1"
    
    # Start frontend development server
    npm run dev
}

# Function to cleanup background processes
cleanup() {
    log "Shutting down development servers..."
    # Kill all background jobs
    jobs -p | xargs -r kill 2>/dev/null || true
    exit 0
}

# Setup signal handling for graceful shutdown
trap cleanup SIGINT SIGTERM EXIT

# Run servers based on options
if [[ "$FRONTEND_ONLY" == "true" ]]; then
    run_frontend
elif [[ "$BACKEND_ONLY" == "true" ]]; then
    run_backend
else
    # Run both servers
    log "Starting both backend and frontend servers..."
    
    # Check if we can run in parallel
    if command -v parallel &> /dev/null; then
        debug "Using GNU parallel to run servers"
        parallel --line-buffer --tagstring '{#}' ::: \
            "run_backend" \
            "run_frontend"
    else
        # Run backend in background, frontend in foreground
        debug "Running backend in background, frontend in foreground"
        
        # Start backend in background
        (
            backend_log "Starting in background..."
            run_backend
        ) &
        BACKEND_PID=$!
        
        # Give backend time to start
        sleep 2
        
        # Start frontend in foreground
        run_frontend
        
        # Wait for background backend process
        wait $BACKEND_PID
    fi
fi

log "Development servers stopped."