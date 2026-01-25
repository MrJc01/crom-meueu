# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies (if any)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Disabling CGO for a static binary (if possible) or keeping it default.
# Go 1.24 usually handles CGO fine, but for alpine->alpine it's okay.
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/server .

# Copy database migration files
COPY --from=builder /app/db ./db

# Copy frontend static files (required for the file server)
COPY --from=builder /app/frontend ./frontend

# Expose the API port
EXPOSE 8080

# Environment variables (Defaults, can be overridden)
ENV SERVER_PORT=8080

# Run the server
CMD ["./server"]
