# --- Stage 1: Build ---
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the rest of the code
COPY . .

# Build the Go app with CGO disabled for static binary
# Build from the `cmd/api` package where main.go lives
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/api

# --- Stage 2: Run ---
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the app port
EXPOSE 8080

# Run the binary
CMD ["./main"]