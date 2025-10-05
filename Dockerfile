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

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

# Copy backend source code
WORKDIR /app
COPY backend/ ./backend/

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/web/dist ./backend/internal/web/dist

# Build the application
WORKDIR /app/backend
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/openlist-strm ./cmd/server

# Stage 3: Final runtime image
FROM alpine:latest

LABEL maintainer="konghang <yslao@outlook.com>"
LABEL description="OpenList-STRM - STRM file generator for Alist"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/bin/openlist-strm .

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/configs && \
    chown -R app:app /app

# Switch to app user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
ENTRYPOINT ["/app/openlist-strm"]
CMD ["-config", "/app/configs/config.yaml"]
