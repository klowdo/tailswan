# Multi-stage Dockerfile for TailSwan
# Bridges strongSwan/swanctl (IPsec) and Tailscale networks

# Build stage for Tailscale
FROM golang:1.23-alpine3.19 AS tailscale-builder

ARG TAILSCALE_VERSION=v1.78.3

RUN apk add --no-cache git

WORKDIR /build

# Clone and build Tailscale
RUN git clone --depth=1 --branch=${TAILSCALE_VERSION} https://github.com/tailscale/tailscale.git && \
    cd tailscale && \
    go mod download && \
    CGO_ENABLED=0 go build -o /tailscale ./cmd/tailscale && \
    CGO_ENABLED=0 go build -o /tailscaled ./cmd/tailscaled

# Runtime stage
FROM alpine:3.19

LABEL org.opencontainers.image.title="TailSwan"
LABEL org.opencontainers.image.description="Bridge strongSwan/swanctl IPsec VPN and Tailscale networks"
LABEL org.opencontainers.image.source="https://github.com/tailswan/tailswan"

# Install strongSwan and required utilities
RUN apk add --no-cache \
    strongswan \
    strongswan-swanctl \
    iptables \
    ip6tables \
    iproute2 \
    curl \
    bash \
    ca-certificates \
    openssl \
    && rm -rf /var/cache/apk/*

# Copy Tailscale binaries from builder
COPY --from=tailscale-builder /tailscale /usr/local/bin/tailscale
COPY --from=tailscale-builder /tailscaled /usr/local/bin/tailscaled

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
    /etc/swanctl/pkcs12 \
    /tailswan

# Copy scripts
COPY scripts/ /tailswan/

# Make scripts executable
RUN chmod +x /tailswan/*.sh

# Enable IP forwarding (will be set in entrypoint for persistence)
RUN echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf && \
    echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf

# Health check
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s --retries=3 \
    CMD /tailswan/healthcheck.sh

ENTRYPOINT ["/tailswan/entrypoint.sh"]
