# Multi-stage build for smaller production image
FROM golang:1.23-alpine AS builder

# Install build dependencies including gcc for CGO (needed for SQLite)
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o main ./cmd/server

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create app directory, data directory, and user
RUN mkdir /app && \
    mkdir -p /data && \
    addgroup -g 1001 app && \
    adduser -D -s /bin/sh -u 1001 -G app app && \
    chown -R app:app /data

# Set working directory
WORKDIR /app

# Copy binary and static files from builder
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

# Change ownership to app user
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port (fly.io expects 8080 by default)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]