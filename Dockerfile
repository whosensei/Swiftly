# --- Stage 1: Build ---
    FROM golang:1.24.5-alpine AS builder

    # Install dependencies
    RUN apk add --no-cache git
    
    # Set working directory
    WORKDIR /app
    
    # Copy go.mod and go.sum first (for caching)
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the code
    COPY . .
    
    # Build the Go app with CGO disabled for static binary
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./main.go
    
    
    # --- Stage 2: Run ---
    FROM alpine:latest
    
    # Install ca-certificates for HTTPS requests
    RUN apk --no-cache add ca-certificates
    
    # Create non-root user
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    
    # Create a directory for the app
    WORKDIR /app
    
    # Copy only the binary from builder
    COPY --from=builder /app/main .
    
    # Change ownership and switch to non-root user
    RUN chown -R appuser:appgroup /app
    USER appuser
    
    # Expose the app port
    EXPOSE 8080
    
    # Run the binary
    CMD ["./main"]