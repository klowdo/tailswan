#!/bin/bash
set -e

echo "============================================"
echo "TailSwan - strongSwan + Tailscale Bridge"
echo "============================================"

# Configuration
TS_STATE_DIR=${TS_STATE_DIR:-/var/lib/tailscale}
TS_SOCKET=${TS_SOCKET:-/var/run/tailscale/tailscaled.sock}
TS_HOSTNAME=${TS_HOSTNAME:-tailswan}
TS_AUTHKEY=${TS_AUTHKEY:-}
TS_ROUTES=${TS_ROUTES:-}
TS_EXTRA_ARGS=${TS_EXTRA_ARGS:-}
TS_SSH=${TS_SSH:-true}
SWAN_CONFIG=${SWAN_CONFIG:-/etc/swanctl/swanctl.conf}
SWAN_AUTO_START=${SWAN_AUTO_START:-false}
SWAN_CONNECTIONS=${SWAN_CONNECTIONS:-}

# Enable IP forwarding
echo "Enabling IP forwarding..."
sysctl -w net.ipv4.ip_forward=1 > /dev/null
sysctl -w net.ipv6.conf.all.forwarding=1 > /dev/null
sysctl -w net.ipv4.conf.all.send_redirects=0 > /dev/null
sysctl -w net.ipv4.conf.default.send_redirects=0 > /dev/null

# Configure iptables for NAT if needed
echo "Setting up iptables rules..."
iptables -t nat -A POSTROUTING -o tailscale0 -j MASQUERADE 2>/dev/null || true
ip6tables -t nat -A POSTROUTING -o tailscale0 -j MASQUERADE 2>/dev/null || true

# Start strongSwan/charon daemon
echo "Starting strongSwan charon daemon..."
mkdir -p /var/run/charon
ipsec start --nofork &
IPSEC_PID=$!

# Wait for charon to be ready
sleep 2

# Load swanctl configuration if it exists
if [ -f "$SWAN_CONFIG" ]; then
    echo "Loading swanctl configuration from $SWAN_CONFIG..."
    swanctl --load-all
else
    echo "Warning: No swanctl configuration found at $SWAN_CONFIG"
fi

# Auto-start IPsec connections if configured
if [ "$SWAN_AUTO_START" = "true" ] && [ -n "$SWAN_CONNECTIONS" ]; then
    echo "Auto-starting IPsec connections: $SWAN_CONNECTIONS"
    IFS=',' read -ra CONNS <<< "$SWAN_CONNECTIONS"
    for conn in "${CONNS[@]}"; do
        echo "Initiating connection: $conn"
        swanctl --initiate --child "$conn" || echo "Warning: Failed to initiate $conn"
    done
fi

# Start Tailscale daemon
echo "Starting Tailscaled..."
tailscaled --state=${TS_STATE_DIR}/tailscaled.state --socket=${TS_SOCKET} --tun=userspace-networking &
TAILSCALED_PID=$!

# Wait for tailscaled to be ready
echo "Waiting for tailscaled to be ready..."
for i in {1..30}; do
    if tailscale status >/dev/null 2>&1; then
        echo "Tailscaled is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: tailscaled failed to start"
        exit 1
    fi
    sleep 1
done

# Build Tailscale up arguments
TS_UP_ARGS="--hostname=${TS_HOSTNAME}"

if [ -n "$TS_AUTHKEY" ]; then
    TS_UP_ARGS="${TS_UP_ARGS} --authkey=${TS_AUTHKEY}"
fi

if [ -n "$TS_ROUTES" ]; then
    echo "Advertising routes to Tailscale: $TS_ROUTES"
    TS_UP_ARGS="${TS_UP_ARGS} --advertise-routes=${TS_ROUTES}"
fi

if [ "$TS_SSH" = "true" ]; then
    echo "Enabling Tailscale SSH..."
    TS_UP_ARGS="${TS_UP_ARGS} --ssh"
fi

if [ -n "$TS_EXTRA_ARGS" ]; then
    TS_UP_ARGS="${TS_UP_ARGS} ${TS_EXTRA_ARGS}"
fi

# Bring up Tailscale
echo "Bringing up Tailscale..."
echo "Command: tailscale up ${TS_UP_ARGS}"
tailscale up ${TS_UP_ARGS}

echo ""
echo "============================================"
echo "TailSwan is running!"
echo "============================================"
echo "Tailscale status:"
tailscale status
echo ""
echo "strongSwan status:"
swanctl --list-conns
echo ""
echo "To initiate an IPsec connection, SSH into this container via Tailscale and run:"
echo "  swanctl --initiate --child <connection-name>"
echo "============================================"

# Monitor both processes
trap 'kill $TAILSCALED_PID $IPSEC_PID 2>/dev/null; exit 0' SIGTERM SIGINT

# Wait for either process to exit
wait -n $TAILSCALED_PID $IPSEC_PID
EXIT_CODE=$?

# If one exits, kill the other
kill $TAILSCALED_PID $IPSEC_PID 2>/dev/null

exit $EXIT_CODE
