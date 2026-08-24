# Multi-stage Dockerfile for the Go backend API server
# The frontend runs in its own container (see Dockerfile.frontend).

# Stage 1: Build the backend binary
FROM golang:1.25-alpine AS builder

# Install build dependencies for SQLite3 (CGO required)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build backend server with CGO enabled for SQLite3
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-extldflags=-static" \
    -o /bin/server \
    ./cmd/server/main.go

# Stage 2: Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /bin/server /app/server

# Copy application assets
COPY --chown=appuser:appuser frontend/ /app/frontend/
COPY --chown=appuser:appuser db/migrations/ /app/db/migrations/
COPY --chown=appuser:appuser db/seeds/ /app/db/seeds/
COPY --chown=appuser:appuser cmd/client/data/ /app/cmd/client/data/
COPY --chown=appuser:appuser certs/ /app/certs/
COPY --chown=appuser:appuser go.mod /app/go.mod

# Create directories for persistent data
RUN mkdir -p /app/db/data /app/frontend/static/images/uploads && \
    chown -R appuser:appuser /app/db/data /app/frontend/static/images/uploads

# Switch to non-root user
USER appuser

# Set default environment variables
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    SERVER_ENVIRONMENT=production \
    SERVER_API_CONTEXT_V1=/api/v1 \
    DB_DRIVER=sqlite3 \
    DB_PATH=db/data/forum.db \
    DB_MIGRATE_ON_START=true \
    DB_SEED_ON_START=false

# Expose port
EXPOSE 8080

# Health check (TLS is configured via SERVER_TLS_* env at runtime)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --no-check-certificate https://localhost:8080/api/v1/health || exit 1

# Run the backend API server directly
CMD ["/app/server"]
