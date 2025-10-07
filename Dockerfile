# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copy frontend files
COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Install build dependencies (gcc and musl-dev required for CGO/SQLite)
RUN apk add --no-cache git make gcc musl-dev

# Copy go mod files
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy backend source code
WORKDIR /app
COPY backend/ ./backend/

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/web/dist ./backend/internal/web/dist

# Build the application with build cache
WORKDIR /app/backend
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o /app/bin/openlist-strm ./cmd/server

# Stage 3: Final runtime image
FROM alpine:latest

LABEL maintainer="konghang <yslao@outlook.com>"
LABEL description="OpenList-STRM - STRM file generator for Alist"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/bin/openlist-strm .

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run via entrypoint script
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/app/openlist-strm", "-config", "/app/configs/config.yaml"]
