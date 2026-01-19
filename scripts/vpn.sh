#!/bin/bash

CONNECTION_NAME="worldstream"

case "$1" in
    up)
        echo "Starting VPN connection: $CONNECTION_NAME"
        swanctl --initiate --child $CONNECTION_NAME
        ;;
    down)
        echo "Stopping VPN connection: $CONNECTION_NAME"
        swanctl --terminate --ike $CONNECTION_NAME
        ;;
    status)
        echo "VPN Status:"
        swanctl --list-sas
        ;;
    list)
        echo "Available connections:"
        swanctl --list-conns
        ;;
    *)
        echo "Usage: $0 {up|down|status|list}"
        echo ""
        echo "Commands:"
        echo "  up      - Initiate VPN connection"
        echo "  down    - Terminate VPN connection"
        echo "  status  - Show active connections"
        echo "  list    - List configured connections"
        exit 1
        ;;
esac
