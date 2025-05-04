FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /govisual-multistorage ./cmd/examples/multistorage

# Use a minimal alpine image for the final container
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /govisual-multistorage .

# Set the entrypoint
ENTRYPOINT ["/app/govisual-multistorage"]

# Expose port 8080
EXPOSE 8080 