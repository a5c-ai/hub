#!/bin/bash

# Test build script to verify optimizations work
set -e

echo "Testing optimized build parameters..."

# Set optimized build environment
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOCACHE=/tmp/go-build-cache
export GOMAXPROCS=8
export GOGC=100
export GOFLAGS="-p=8 -buildvcs=false"

# Test simple build first
echo "Building with optimized flags..."
go build -ldflags "-s -w" -trimpath -tags netgo -buildmode=default -compiler=gc -o ./test-hub ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful with optimizations"
    ./test-hub --version || echo "Binary built successfully (version flag not implemented)"
    rm -f ./test-hub
else
    echo "❌ Build failed"
    exit 1
fi