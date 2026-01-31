# STAGE 1: Build the binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first to leverage Docker's caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp main.go

# STAGE 2: Create the production image
FROM alpine:3.21

# Add a non-root user for security
RUN adduser -D appuser
USER appuser

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /myapp /myapp

# Expose the gRPC port
EXPOSE 50051

# Run the application
ENTRYPOINT ["/myapp"]
