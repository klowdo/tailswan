#!/bin/bash
# TailSwan status and management helper script

set -e

command="${1:-status}"

case "$command" in
    status)
        echo "============================================"
        echo "TailSwan Status"
        echo "============================================"
        echo ""
        echo "--- Tailscale Status ---"
        tailscale status
        echo ""
        echo "--- IPsec Connections ---"
        swanctl --list-conns
        echo ""
        echo "--- Active SAs (Security Associations) ---"
        swanctl --list-sas
        echo ""
        echo "--- Routing Table ---"
        ip route show
        echo ""
        ;;

    connections|conns)
        echo "Available IPsec connections:"
        swanctl --list-conns
        ;;

    sas)
        echo "Active Security Associations:"
        swanctl --list-sas
        ;;

    start)
        if [ -z "$2" ]; then
            echo "Usage: $0 start <connection-name>"
            exit 1
        fi
        echo "Initiating IPsec connection: $2"
        swanctl --initiate --child "$2"
        ;;

    stop)
        if [ -z "$2" ]; then
            echo "Usage: $0 stop <connection-name>"
            exit 1
        fi
        echo "Terminating IPsec connection: $2"
        swanctl --terminate --ike "$2"
        ;;

    reload)
        echo "Reloading swanctl configuration..."
        swanctl --load-all
        ;;

    routes)
        echo "--- IP Routes ---"
        ip route show
        echo ""
        echo "--- Tailscale Routes ---"
        tailscale status | grep -A 100 "Advertised routes:" || echo "No routes advertised"
        ;;

    help|--help|-h)
        echo "TailSwan Management Script"
        echo ""
        echo "Usage: $0 [command] [arguments]"
        echo ""
        echo "Commands:"
        echo "  status              Show complete TailSwan status (default)"
        echo "  connections, conns  List IPsec connections"
        echo "  sas                 List active Security Associations"
        echo "  start <name>        Initiate an IPsec connection"
        echo "  stop <name>         Terminate an IPsec connection"
        echo "  reload              Reload swanctl configuration"
        echo "  routes              Show routing information"
        echo "  help                Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 status"
        echo "  $0 start site-to-site"
        echo "  $0 stop site-to-site"
        ;;

    *)
        echo "Unknown command: $command"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac
