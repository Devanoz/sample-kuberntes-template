# STAGE 1: Build the binary
# Using the official Go 1.25+ image as the builder
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker's caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp main.go

# STAGE 2: Create the production image
# Using a minimal image (scratch or alpine) to keep the size down
FROM alpine:3.21

# Add a non-root user for security
RUN adduser -D appuser
USER appuser

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /myapp /myapp

# Expose the port your app listens on
EXPOSE 3000

# Run the application
ENTRYPOINT ["/myapp"]
