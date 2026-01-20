package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "healthcheck":
			runHealthCheck()
		case "status":
			runStatus()
		case "connections", "conns":
			runConnections()
		case "sas":
			runSAs()
		case "start":
			runStart(os.Args[2:])
		case "stop":
			runStop(os.Args[2:])
		case "reload":
			runReload()
		case "help", "-h", "--help":
			printHelp()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "Run 'tailswan help' for usage information")
			os.Exit(1)
		}
		return
	}

	runSupervisor()
}

func printHelp() {
	fmt.Println("TailSwan - IPsec & Tailscale VPN Supervisor")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  tailswan                    Run supervisor (start all services)")
	fmt.Println("  tailswan healthcheck        Check if all services are healthy")
	fmt.Println("  tailswan status             Show status of all services")
	fmt.Println("  tailswan connections        List strongSwan connections")
	fmt.Println("  tailswan sas                List security associations")
	fmt.Println("  tailswan start <conn>       Initiate a connection")
	fmt.Println("  tailswan stop <conn>        Terminate a connection")
	fmt.Println("  tailswan reload             Reload strongSwan configuration")
	fmt.Println("  tailswan help               Show this help message")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  CONTROL_PORT               Control server port (default: 8080)")
	fmt.Println("  USE_TSNET                  Use tsnet instead of Tailscale Serve (true/false)")
	fmt.Println("  TS_STATE_DIR               Tailscale state directory")
	fmt.Println("  TS_SOCKET                  Tailscale socket path")
	fmt.Println("  TS_HOSTNAME                Tailscale hostname")
	fmt.Println("  TS_AUTHKEY                 Tailscale auth key")
	fmt.Println("  TS_ROUTES                  Tailscale routes (comma-separated)")
	fmt.Println("  TS_SSH                     Enable Tailscale SSH (true/false)")
	fmt.Println("  TS_EXTRA_ARGS              Extra tailscale up arguments")
	fmt.Println("  SWAN_CONFIG                Path to swanctl.conf")
	fmt.Println("  SWAN_AUTO_START            Auto-start connections (true/false)")
	fmt.Println("  SWAN_CONNECTIONS           Connections to auto-start (comma-separated)")
}
