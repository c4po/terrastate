# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set build-time variables for versioning
ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

# Set environment variables for the application
ENV VERSION=$VERSION
ENV GIT_SHA=$GIT_SHA
ENV BUILD_TIME=$BUILD_TIME

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with versioning information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION} -X main.GitSha=${GIT_SHA} -X main.BuildTime=${BUILD_TIME}" -o terraform-state-backend ./cmd/server

# Final stage
FROM alpine:3.19

# Install CA certificates for HTTPS
# RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/terraform-state-backend .

# Default environment variables
ENV STORAGE_TYPE=s3
ENV PORT=8080
ENV VERSION=${VERSION}
ENV GIT_SHA=${GIT_SHA}
ENV BUILD_TIME=${BUILD_TIME}

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./terraform-state-backend"] 
