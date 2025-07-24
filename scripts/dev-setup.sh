#!/bin/bash

set -e

echo "Setting up Hub development environment..."

# Install dependencies
echo "Installing Go dependencies..."
go mod download
go mod tidy

# Create config file if it doesn't exist
if [ ! -f "config.yaml" ]; then
    echo "Creating development config file..."
    cp config.example.yaml config.yaml
    echo "✅ Created config.yaml from example. Please update with your settings."
else
    echo "✅ config.yaml already exists"
fi

# Check if PostgreSQL is running (optional)
if command -v psql &> /dev/null; then
    echo "📋 PostgreSQL client found. Make sure PostgreSQL server is running."
    echo "   Default connection: postgresql://hub:password@localhost:5432/hub"
    echo "   You can update database settings in config.yaml"
fi

# Build the application
echo "Building the application..."
go build -o hub ./cmd/server

# Run tests
echo "Running tests..."
go test ./...

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "To start the server:"
echo "  ./hub"
echo ""
echo "To run with custom config:"
echo "  export ENVIRONMENT=development"
echo "  export DB_HOST=localhost"
echo "  export DB_USER=hub" 
echo "  export DB_PASSWORD=password"
echo "  export DB_NAME=hub"
echo "  ./hub"
echo ""
echo "Health check endpoint: http://localhost:8080/health"
echo "API endpoints: http://localhost:8080/api/v1/"
