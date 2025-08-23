# Official Docker image for cariddi
# Build stage
FROM golang:1.24-alpine AS builder

# Set build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.commit=${VCS_REF}" \
    -a -installsuffix cgo \
    -o cariddi \
    ./cmd/cariddi

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1001 -S cariddi && \
    adduser -S -D -H -u 1001 -h /home/cariddi -s /sbin/nologin -G cariddi -g cariddi cariddi

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/cariddi /usr/local/bin/cariddi

# Make binary executable
RUN chmod +x /usr/local/bin/cariddi

# Switch to non-root user
USER cariddi

# Add metadata
LABEL maintainer="edoardottt <edoardott@gmail.com>" \
      description="Take a list of domains, crawl urls and scan for endpoints, secrets, api keys, file extensions, tokens and more" \
      version="${VERSION}" \
      url="https://github.com/edoardottt/cariddi" \
      vcs-url="https://github.com/edoardottt/cariddi" \
      vcs-ref="${VCS_REF}" \
      build-date="${BUILD_DATE}"

# Set default entrypoint
ENTRYPOINT ["cariddi"]

# Default command (show help)
CMD ["--help"]
