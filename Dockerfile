# Build stage
FROM golang:1.23-alpine AS builder

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
# Set build-time variables for versioning
ARG VERSION=dev
ARG GIT_SHA=none

# Set environment variables for the application
ENV VERSION=$VERSION
ENV GIT_SHA=$GIT_SHA

# Build the application
RUN make build-linux

# Final stage
FROM alpine:3.19

# Install CA certificates for HTTPS
# RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/build/linux_amd64/terraform-state-backend .

# Default environment variables
ENV STORAGE_TYPE=s3
ENV PORT=8080
ENV VERSION=${VERSION}
ENV GIT_SHA=${GIT_SHA}

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./terraform-state-backend"] 
