# Multi-stage Dockerfile for TailSwan
# Bridges strongSwan/swanctl (IPsec) and Tailscale networks
#
# Build arguments:
#   GO_VERSION        - Go version for build stages (default: 1.25.6)
#   ALPINE_VERSION    - Alpine Linux version (default: 3.22)
#   TAILSCALE_VERSION - Tailscale version tag (default: latest)
#
# Example:
#   docker build --build-arg GO_VERSION=1.25.5 --build-arg TAILSCALE_VERSION=v1.92.5 .

ARG GO_VERSION=1.25.6
ARG ALPINE_VERSION=3.22
ARG TAILSCALE_VERSION=v1.92.5

# Base builder stage with dependencies
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base-builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build stage for TailSwan supervisor
FROM base-builder AS supervisor-builder

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tailswan ./cmd/tailswan

# Build stage for TailSwan control server
FROM base-builder AS controlserver-builder

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /controlserver ./cmd/controlserver

# Runtime stage - use Tailscale image as base
ARG TAILSCALE_VERSION
FROM ghcr.io/tailscale/tailscale:${TAILSCALE_VERSION}

LABEL org.opencontainers.image.title="TailSwan"
LABEL org.opencontainers.image.description="Bridge strongSwan/swanctl IPsec VPN and Tailscale networks"
LABEL org.opencontainers.image.source="https://github.com/tailswan/tailswan"

# Install strongSwan and bash on top of Tailscale image
# (Tailscale image already includes iptables, iproute2, ca-certificates, and legacy iptables symlinks)
RUN apk add --no-cache \
    strongswan \
    bash \
    && rm -rf /var/cache/apk/*

# Copy TailSwan binaries from builders
COPY --from=supervisor-builder /tailswan /usr/local/bin/tailswan
COPY --from=controlserver-builder /controlserver /usr/local/bin/controlserver

# Install shell completions
RUN mkdir -p /etc/bash_completion.d \
    /usr/share/zsh/site-functions \
    /usr/share/fish/vendor_completions.d && \
    tailswan completion bash > /etc/bash_completion.d/tailswan && \
    tailswan completion zsh > /usr/share/zsh/site-functions/_tailswan && \
    tailswan completion fish > /usr/share/fish/vendor_completions.d/tailswan.fish

# Copy and install welcome banner
COPY assets/banner.txt /etc/motd

# Create necessary directories
RUN mkdir -p /var/run/tailscale \
    /var/lib/tailscale \
    /etc/swanctl/conf.d \
    /etc/swanctl/x509 \
    /etc/swanctl/x509ca \
    /etc/swanctl/x509crl \
    /etc/swanctl/pubkey \
    /etc/swanctl/private \
    /etc/swanctl/rsa \
    /etc/swanctl/ecdsa \
    /etc/swanctl/pkcs12

# Set environment variable for tailscale CLI to find the socket
ENV TAILSCALE_SOCKET=/var/run/tailscale/tailscaled.sock

# Health check
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s --retries=3 \
    CMD tailswan healthcheck

ENTRYPOINT ["tailswan"]
CMD ["serve"]
