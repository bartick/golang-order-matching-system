# ----- Stage 1: Build -----
FROM golang:1.24.3 AS builder

WORKDIR /app

# Only copy go.mod and go.sum to leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/golang-order-matching-system ./main.go

# ----- Stage 2: Runtime -----
FROM alpine:latest

# Install certs for HTTPS if needed (common for HTTP clients)
RUN apk --no-cache add ca-certificates curl

# Copy the binary from the builder stage
COPY --from=builder /app/bin/golang-order-matching-system /usr/local/bin/golang-order-matching-system

ENV GO_ENV=production

# Set binary as entrypoint
ENTRYPOINT ["/usr/local/bin/golang-order-matching-system"]

# Expose the application port
EXPOSE 8080

# Add a healthcheck endpoint if available
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/ping || exit 1
