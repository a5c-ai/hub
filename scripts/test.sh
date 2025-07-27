#!/bin/bash

set -e

# Configuration
TEST_ENV=${TEST_ENV:-test}
COVERAGE=${COVERAGE:-false}
E2E=${E2E:-false}
UNIT_ONLY=${UNIT_ONLY:-false}
E2E_ONLY=${E2E_ONLY:-false}
BUILD_FIRST=${BUILD_FIRST:-true}
PARALLEL=${PARALLEL:-true}

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

test_log() {
    echo -e "${MAGENTA}[TEST]${NC} $1"
}

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --unit-only            Run only unit tests"
    echo "  --e2e-only             Run only end-to-end tests"
    echo "  --no-build             Skip building before tests"
    echo "  --no-e2e               Skip end-to-end tests"
    echo "  --coverage             Generate coverage reports"
    echo "  --no-parallel          Run tests sequentially"
    echo "  --env ENV              Test environment (default: test)"
    echo "  --help                 Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  TEST_ENV               Test environment"
    echo "  COVERAGE               Generate coverage (true/false)"
    echo "  E2E                    Run E2E tests (true/false)"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            UNIT_ONLY=true
            E2E=false
            shift
            ;;
        --e2e-only)
            E2E_ONLY=true
            shift
            ;;
        --no-build)
            BUILD_FIRST=false
            shift
            ;;
        --no-e2e)
            E2E=false
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --no-parallel)
            PARALLEL=false
            shift
            ;;
        --env)
            TEST_ENV="$2"
            shift 2
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
if [[ "$UNIT_ONLY" == "true" && "$E2E_ONLY" == "true" ]]; then
    error "Cannot specify both --unit-only and --e2e-only"
    exit 1
fi

log "Starting Hub test suite..."
log "Test environment: $TEST_ENV"
log "Coverage: $COVERAGE"
log "E2E tests: $E2E"
log "Build first: $BUILD_FIRST"

# Set test environment variables
# Set test environment variables
export ENVIRONMENT="$TEST_ENV"
export NODE_ENV="test"
export GO_ENV="test"

# Configure test database parameters
export DB_HOST="${TEST_DB_HOST:-localhost}"
export DB_PORT="${TEST_DB_PORT:-5432}"
export DB_NAME="${TEST_DB_NAME:-hub_test}"
export DB_USER="${TEST_DB_USER:-hub}"
export DB_PASSWORD="${TEST_DB_PASSWORD:-password}"

# Setup and teardown for PostgreSQL test container
setup_test_db() {
    test_log "Starting PostgreSQL test container..."
    docker run -d --name hub-test-db -e POSTGRES_PASSWORD=$DB_PASSWORD -e POSTGRES_USER=$DB_USER -e POSTGRES_DB=$DB_NAME -p $DB_PORT:5432 postgres:16
    # Provide password for psql to connect without interactive prompt
    export PGPASSWORD="$DB_PASSWORD"
    test_log "Waiting for PostgreSQL to be ready (timeout: ${GO_TEST_TIMEOUT:-5m})..."
    if ! timeout "${GO_TEST_TIMEOUT:-5m}" bash -c \
        "until psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\\l' &>/dev/null; do sleep 1; done"; then
        error "Timed out waiting for PostgreSQL after ${GO_TEST_TIMEOUT:-5m}"
        exit 1
    fi
    test_log "PostgreSQL test container is ready."
}
cleanup_test_db() {
    test_log "Stopping PostgreSQL test container..."
    docker stop hub-test-db
}
trap cleanup_test_db EXIT

setup_test_db

# Build before testing if requested
if [[ "$BUILD_FIRST" == "true" && "$E2E_ONLY" == "false" ]]; then
    log "Building application before tests..."
    ./scripts/build.sh
fi

# Track test results
UNIT_TESTS_PASSED=true
FRONTEND_TESTS_PASSED=true
E2E_TESTS_PASSED=true

# Function to run Go unit tests
run_go_tests() {
    if [[ ! -f "go.mod" ]]; then
        warn "No go.mod found, skipping Go tests"
        return 0
    fi
    
    test_log "Running Go unit tests..."
    
    # Set up test database if needed
    export DB_HOST="${TEST_DB_HOST:-localhost}"
    export DB_PORT="${TEST_DB_PORT:-5432}"
    export DB_NAME="${TEST_DB_NAME:-hub_test}"
    export DB_USER="${TEST_DB_USER:-hub}"
    export DB_PASSWORD="${TEST_DB_PASSWORD:-password}"
    
    local go_test_flags="-v"
    if [[ "$PARALLEL" == "true" ]]; then
        go_test_flags="$go_test_flags -parallel 4"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_log "Running with coverage..."
        go_test_flags="$go_test_flags -coverprofile=coverage.out -covermode=atomic"
        
        # Run tests with coverage
        if go test $go_test_flags ./...; then
            test_log "Go tests passed ✅"
            
            # Generate coverage report
            go tool cover -html=coverage.out -o coverage.html
            go tool cover -func=coverage.out | tail -1
            test_log "Coverage report generated: coverage.html"
        else
            error "Go tests failed ❌"
            UNIT_TESTS_PASSED=false
        fi
    else
        # Run tests without coverage
        if go test $go_test_flags ./...; then
            test_log "Go tests passed ✅"
        else
            error "Go tests failed ❌"
            UNIT_TESTS_PASSED=false
        fi
    fi
}

# Function to run frontend tests
run_frontend_tests() {
    if [[ ! -d "frontend" || ! -f "frontend/package.json" ]]; then
        warn "Frontend directory not found, skipping frontend tests"
        return 0
    fi
    
    test_log "Running frontend tests..."
    
    cd frontend
    
    # Ensure dependencies are installed
    if [[ ! -d "node_modules" ]]; then
        test_log "Installing frontend dependencies..."
        npm ci
    fi
    
    local npm_test_cmd="npm run test"
    if [[ "$COVERAGE" == "true" ]]; then
        npm_test_cmd="npm run test:ci"
    fi
    
    # Set test environment
    export NODE_ENV="test"
    export CI=true  # Prevent Jest from watching files
    
    # Run frontend tests
    if $npm_test_cmd; then
        test_log "Frontend tests passed ✅"
    else
        error "Frontend tests failed ❌"
        FRONTEND_TESTS_PASSED=false
    fi
    
    cd ..
}

# Function to run end-to-end tests
run_e2e_tests() {
    if [[ ! -d "frontend" || ! -f "frontend/package.json" ]]; then
        warn "Frontend directory not found, skipping E2E tests"
        return 0
    fi
    
    test_log "Running end-to-end tests..."
    
	cd frontend
	
    # Install Playwright browsers
    test_log "Installing Playwright browsers..."
    npx playwright install --with-deps

    # Start test servers for E2E tests
    test_log "Starting test servers for E2E tests..."

    # Start backend in background
    (
        export ENVIRONMENT="test"
        export PORT="8081"
        export DB_NAME="hub_test"
        export DB_HOST="${TEST_DB_HOST:-localhost}"
        export DB_PORT="${TEST_DB_PORT:-5432}"
        export DB_USER="${TEST_DB_USER:-hub}"
        export DB_PASSWORD="${TEST_DB_PASSWORD:-password}"
        cd ..
        go run ./cmd/server
    ) &
    BACKEND_PID=$!

    # Start frontend in background
    (
        export NODE_ENV="test"
        export PORT="3001"
        export NEXT_PUBLIC_API_URL="http://localhost:8081"
        npm run dev
    ) &
    FRONTEND_PID=$!

    # Wait for servers to start
    sleep 5

    # Setup cleanup on exit
    cleanup_servers() {
        test_log "Stopping test servers..."
        kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    }
    trap cleanup_servers EXIT

    # Check if E2E tests are configured
    if ! npm run --silent test:e2e --dry-run &>/dev/null; then
        warn "E2E tests not configured, skipping"
        cd ..
        return 0
    fi

    # Run E2E tests
    export E2E_BASE_URL="${E2E_BASE_URL:-http://localhost:3001}"
    export E2E_API_URL="${E2E_API_URL:-http://localhost:8081}"
    if npm run test:e2e; then
        test_log "E2E tests passed ✅"
    else
        error "E2E tests failed ❌"
        E2E_TESTS_PASSED=false
    fi
    cd ..
}

# Function to run linting
run_linting() {
    test_log "Running code quality checks..."
    
    # Go linting
    if command -v golangci-lint &> /dev/null && [[ -f "go.mod" ]]; then
        test_log "Running Go linting..."
        if golangci-lint run; then
            test_log "Go linting passed ✅"
        else
            warn "Go linting issues found ⚠️"
        fi
    fi
    
    # Frontend linting
    if [[ -d "frontend" ]]; then
        cd frontend
        if npm run --silent lint --dry-run &>/dev/null; then
            test_log "Running frontend linting..."
            if npm run lint; then
                test_log "Frontend linting passed ✅"
            else
                warn "Frontend linting issues found ⚠️"
            fi
        fi
        cd ..
    fi
}

# Run tests based on options
if [[ "$E2E_ONLY" == "true" ]]; then
    run_e2e_tests
elif [[ "$UNIT_ONLY" == "true" ]]; then
    run_go_tests
    run_frontend_tests
    run_linting
else
    # Run all tests
    log "Running all tests..."
    
    # Run unit tests first
    run_go_tests
    run_frontend_tests
    run_linting
    
    # Run E2E tests if enabled
    if [[ "$E2E" == "true" ]]; then
        run_e2e_tests
    fi
fi

# Report results
log "Test Results Summary:"
log "===================="

if [[ "$UNIT_ONLY" == "false" && "$E2E_ONLY" == "false" ]]; then
    log "Go unit tests: $([ "$UNIT_TESTS_PASSED" == "true" ] && echo "✅ PASSED" || echo "❌ FAILED")"
    log "Frontend tests: $([ "$FRONTEND_TESTS_PASSED" == "true" ] && echo "✅ PASSED" || echo "❌ FAILED")"
    if [[ "$E2E" == "true" ]]; then
        log "E2E tests: $([ "$E2E_TESTS_PASSED" == "true" ] && echo "✅ PASSED" || echo "❌ FAILED")"
    fi
elif [[ "$UNIT_ONLY" == "true" ]]; then
    log "Unit tests: $([ "$UNIT_TESTS_PASSED" == "true" ] && [ "$FRONTEND_TESTS_PASSED" == "true" ] && echo "✅ PASSED" || echo "❌ FAILED")"
else
    log "E2E tests: $([ "$E2E_TESTS_PASSED" == "true" ] && echo "✅ PASSED" || echo "❌ FAILED")"
fi

# Exit with error if any tests failed
if [[ "$UNIT_TESTS_PASSED" == "false" || "$FRONTEND_TESTS_PASSED" == "false" || "$E2E_TESTS_PASSED" == "false" ]]; then
    error "Some tests failed!"
    exit 1
fi

log "🎉 All tests passed successfully!"
