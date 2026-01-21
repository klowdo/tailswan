# Multi-stage Dockerfile for TailSwan
# Bridges strongSwan/swanctl (IPsec) and Tailscale networks

# Build stage for Tailscale
FROM golang:1.25.6-alpine3.22 AS tailscale-builder

ARG TAILSCALE_VERSION=v1.92.5


RUN apk add --no-cache git

WORKDIR /build

# Clone and build Tailscale
RUN git config --global advice.detachedHead false
RUN git clone  --single-branch  --branch=${TAILSCALE_VERSION} https://github.com/tailscale/tailscale.git && \
    cd tailscale && \
    go mod download && \
    CGO_ENABLED=0 go build -o /tailscale ./cmd/tailscale && \
    CGO_ENABLED=0 go build -o /tailscaled ./cmd/tailscaled

# Base builder stage with dependencies
FROM golang:1.25.6-alpine3.22 AS base-builder

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

# Runtime stage
FROM alpine:3.22

LABEL org.opencontainers.image.title="TailSwan"
LABEL org.opencontainers.image.description="Bridge strongSwan/swanctl IPsec VPN and Tailscale networks"
LABEL org.opencontainers.image.source="https://github.com/tailswan/tailswan"

# Install strongSwan and required utilities
RUN apk add --no-cache \
    strongswan \
    iptables \
    ip6tables \
    iproute2 \
    curl \
    ca-certificates \
    openssl \
    && rm -rf /var/cache/apk/*

# Alpine 3.19+ replaced legacy iptables with nftables based implementation.
# Tailscale and strongSwan are used on some hosts that don't support nftables,
# such as Synology NAS, so link iptables back to legacy version.
# Hosts that don't require legacy iptables should be able to use in nftables mode.
# See https://github.com/tailscale/tailscale/issues/17854
RUN rm /usr/sbin/iptables && ln -s /usr/sbin/iptables-legacy /usr/sbin/iptables && \
    rm /usr/sbin/ip6tables && ln -s /usr/sbin/ip6tables-legacy /usr/sbin/ip6tables

# Copy Tailscale binaries from builder
COPY --from=tailscale-builder /tailscale /usr/local/bin/tailscale
COPY --from=tailscale-builder /tailscaled /usr/local/bin/tailscaled

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

# Health check
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s --retries=3 \
    CMD tailswan healthcheck

ENTRYPOINT ["tailswan"]
CMD ["serve"]
