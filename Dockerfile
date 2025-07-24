# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates (required for some Go modules)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and git (needed for git operations)
RUN apk --no-cache add ca-certificates git

# Create non-root user first
RUN addgroup -g 1001 -S hub && \
    adduser -S hub -u 1001 -G hub

# Create necessary directories
RUN mkdir -p /repositories /app && \
    chown -R hub:hub /repositories /app

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Switch to non-root user
USER hub

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]