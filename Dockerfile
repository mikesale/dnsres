# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/dnsres .

# Copy config file
COPY config.json .

# Create log directory
RUN mkdir -p logs

# Expose ports
EXPOSE 8080 9090

# Set environment variables
ENV TZ=UTC

# Run the application
CMD ["./dnsres"] 