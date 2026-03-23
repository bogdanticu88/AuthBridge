# Use Go 1.25 as the builder
FROM golang:1.25-bookworm AS builder

# Set working directory
WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=1 go build -o /authbridge main.go

# Use a slim Debian image for runtime
FROM debian:bookworm-slim

# Install CA certificates for HTTPS token refresh
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder
COPY --from=builder /authbridge /usr/local/bin/authbridge

# Create a non-root user and home directory for .authbridge
RUN useradd -m authbridge
USER authbridge
WORKDIR /home/authbridge

# Expose the API port
EXPOSE 9999

# Set the command to start the daemon
ENTRYPOINT ["authbridge", "start", "--host", "0.0.0.0"]
