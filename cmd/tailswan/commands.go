package main

import (
	"fmt"
	"log/slog"
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
		slog.Error("Failed to list connections", "error", err)
		os.Exit(1)
	}
}

func runSAs() {
	sw := &supervisor.SwanService{}
	if err := sw.ListSAs(); err != nil {
		slog.Error("Failed to list SAs", "error", err)
		os.Exit(1)
	}
}

func runStart(args []string) {
	if len(args) == 0 {
		slog.Error("Connection name required")
		fmt.Fprintln(os.Stderr, "Usage: tailswan start <connection>")
		os.Exit(1)
	}

	conn := args[0]
	sw := &supervisor.SwanService{}
	if err := sw.Initiate(conn); err != nil {
		slog.Error("Failed to start connection", "connection", conn, "error", err)
		os.Exit(1)
	}
	slog.Info("Connection initiated", "connection", conn)
}

func runStop(args []string) {
	if len(args) == 0 {
		slog.Error("Connection name required")
		fmt.Fprintln(os.Stderr, "Usage: tailswan stop <connection>")
		os.Exit(1)
	}

	conn := args[0]
	sw := &supervisor.SwanService{}
	if err := sw.Terminate(conn); err != nil {
		slog.Error("Failed to stop connection", "connection", conn, "error", err)
		os.Exit(1)
	}
	slog.Info("Connection terminated", "connection", conn)
}

func runReload() {
	sw := &supervisor.SwanService{}
	if err := sw.Reload(); err != nil {
		slog.Error("Failed to reload configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration reloaded")
}

func runHealthCheck() {
	if err := supervisor.HealthCheck(); err != nil {
		slog.Error("Health check failed", "error", err)
		os.Exit(1)
	}
	slog.Info("All services healthy")
}
