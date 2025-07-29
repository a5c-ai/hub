# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Install dependencies needed for node-gyp
RUN apk add --no-cache python3 make g++

# Copy package files
COPY package*.json ./

# Install all dependencies for build stage
RUN npm ci --legacy-peer-deps && npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM node:18-alpine AS production

WORKDIR /app

# Install only production dependencies (skip prepare scripts)
COPY package*.json ./
RUN npm ci --omit=dev --ignore-scripts --legacy-peer-deps && npm cache clean --force

# Copy built application from builder stage
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./
COPY --from=builder /app/package.json ./

# Create non-root user
RUN addgroup -g 1001 -S hub && \
    adduser -S hub -u 1001 -G hub

# Change ownership of the app directory
RUN chown -R hub:hub /app

# Switch to non-root user
USER hub

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start the application
CMD ["npm", "start"]
