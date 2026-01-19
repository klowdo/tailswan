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
READY=false
for i in {1..60}; do
    STATUS_OUTPUT=$(tailscale status 2>&1 || true)
    STATUS_EXIT=$?

    # Check if tailscale is responding (either logged in or logged out is OK)
    if echo "$STATUS_OUTPUT" | grep -qE "(Logged out|logged in|Health check)"; then
        echo "✓ Tailscaled is ready and responding!"
        echo "  Status: $STATUS_OUTPUT" | head -1
        READY=true
        break
    fi

    # Show progress
    if [ $((i % 10)) -eq 0 ]; then
        echo "  Still waiting for tailscaled... ($i/60 seconds)"
        echo "  Current status output: ${STATUS_OUTPUT:-<no output>}"
    fi

    sleep 1
done

if [ "$READY" = "false" ]; then
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║  CONTAINER RESTART REASON: Tailscaled Failed to Start     ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Tailscaled did not respond within 60 seconds."
    echo ""
    echo "Process status:"
    ps aux | grep -E "(tailscaled|PID)" || echo "  Cannot check processes"
    echo ""
    echo "Socket status:"
    ls -la "$TS_SOCKET" 2>&1 || echo "  Socket not found at $TS_SOCKET"
    echo ""
    echo "Tailscale status output:"
    tailscale status 2>&1 || echo "  No status available"
    echo ""
    echo "This container will now exit and restart..."
    echo "════════════════════════════════════════════════════════════"
    exit 1
fi

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
trap 'echo "Received shutdown signal, cleaning up..."; kill $TAILSCALED_PID $IPSEC_PID 2>/dev/null; exit 0' SIGTERM SIGINT

echo ""
echo "✓ TailSwan is now running and monitoring services..."
echo "  Tailscaled PID: $TAILSCALED_PID"
echo "  IPsec (charon) PID: $IPSEC_PID"
echo ""

# Wait for either process to exit
wait -n $TAILSCALED_PID $IPSEC_PID
EXIT_CODE=$?

# Figure out which process died
TAILSCALED_RUNNING=false
IPSEC_RUNNING=false
kill -0 $TAILSCALED_PID 2>/dev/null && TAILSCALED_RUNNING=true
kill -0 $IPSEC_PID 2>/dev/null && IPSEC_RUNNING=true

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         CONTAINER RESTART REASON: Process Died             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

if [ "$TAILSCALED_RUNNING" = "false" ]; then
    echo "✗ TAILSCALED PROCESS EXITED (PID: $TAILSCALED_PID)"
    echo "  Exit code: $EXIT_CODE"
    echo "  This usually means Tailscale crashed or was terminated."
fi

if [ "$IPSEC_RUNNING" = "false" ]; then
    echo "✗ IPSEC (CHARON) PROCESS EXITED (PID: $IPSEC_PID)"
    echo "  Exit code: $EXIT_CODE"
    echo "  This usually means strongSwan crashed or was terminated."
fi

echo ""
echo "Cleaning up remaining processes..."
kill $TAILSCALED_PID $IPSEC_PID 2>/dev/null

echo "Container will now exit and Docker will restart it..."
echo "════════════════════════════════════════════════════════════"

exit $EXIT_CODE
