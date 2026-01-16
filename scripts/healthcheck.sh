#!/bin/bash

# Health check script for TailSwan

# Check if tailscaled is running
if ! tailscale status >/dev/null 2>&1; then
    echo "UNHEALTHY: tailscaled is not responding"
    exit 1
fi

# Check if charon daemon is running
if ! swanctl --version >/dev/null 2>&1; then
    echo "UNHEALTHY: swanctl/charon is not responding"
    exit 1
fi

# Check if ipsec is running
if ! pgrep -x charon >/dev/null; then
    echo "UNHEALTHY: charon daemon is not running"
    exit 1
fi

echo "HEALTHY: All services running"
exit 0
