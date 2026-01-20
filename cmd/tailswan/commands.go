package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/klowdo/tailswan/internal/supervisor"
)

func runStatus() {
	fmt.Println("=== Tailscale Status ===")
	cmd := exec.Command("tailscale", "status")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println("\n=== strongSwan Connections ===")
	cmd = exec.Command("swanctl", "--list-conns")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func runConnections() {
	sw := &supervisor.SwanService{}
	if err := sw.ListConnections(); err != nil {
		os.Exit(1)
	}
}

func runSAs() {
	sw := &supervisor.SwanService{}
	if err := sw.ListSAs(); err != nil {
		os.Exit(1)
	}
}

func runStart(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: connection name required")
		fmt.Fprintln(os.Stderr, "Usage: tailswan start <connection>")
		os.Exit(1)
	}

	conn := args[0]
	sw := &supervisor.SwanService{}
	if err := sw.Initiate(conn); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection '%s' initiated\n", conn)
}

func runStop(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: connection name required")
		fmt.Fprintln(os.Stderr, "Usage: tailswan stop <connection>")
		os.Exit(1)
	}

	conn := args[0]
	sw := &supervisor.SwanService{}
	if err := sw.Terminate(conn); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connection '%s' terminated\n", conn)
}

func runReload() {
	sw := &supervisor.SwanService{}
	if err := sw.Reload(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration reloaded")
}

func runHealthCheck() {
	if err := supervisor.HealthCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All services healthy")
}
