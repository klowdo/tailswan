package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/supervisor"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *slog.Logger
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tailswan",
	Short: "TailSwan - IPsec & Tailscale VPN Supervisor",
	Long:  `TailSwan is a unified supervisor for managing strongSwan IPsec and Tailscale VPN connections.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg = config.Load()
		initLogger(cfg)
	},
	Run: func(cmd *cobra.Command, args []string) {
		runSupervisor()
	},
}

var healthCheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check if all services are healthy",
	Run: func(cmd *cobra.Command, args []string) {
		runHealthCheck()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all services",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

var connectionsCmd = &cobra.Command{
	Use:     "connections",
	Aliases: []string{"conns"},
	Short:   "List strongSwan connections",
	Run: func(cmd *cobra.Command, args []string) {
		runConnections()
	},
}

var sasCmd = &cobra.Command{
	Use:   "sas",
	Short: "List security associations",
	Run: func(cmd *cobra.Command, args []string) {
		runSAs()
	},
}

var startCmd = &cobra.Command{
	Use:   "start <connection>",
	Short: "Initiate a connection",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStart(args)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <connection>",
	Short: "Terminate a connection",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args)
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload strongSwan configuration",
	Run: func(cmd *cobra.Command, args []string) {
		runReload()
	},
}

func init() {
	rootCmd.AddCommand(
		healthCheckCmd,
		statusCmd,
		connectionsCmd,
		sasCmd,
		startCmd,
		stopCmd,
		reloadCmd,
	)
}

func initLogger(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func runSupervisor() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	supervisorCfg := supervisor.Config{
		ControlPort:       cfg.Port,
		TailscaleStateDir: cfg.Tailscale.StateDir,
		TailscaleSocket:   cfg.Tailscale.Socket,
		UseTsnet:          cfg.Tailscale.UseTsnet,
		TailscaleConfig: supervisor.TailscaleConfig{
			StateDir:  cfg.Tailscale.StateDir,
			Socket:    cfg.Tailscale.Socket,
			Hostname:  cfg.Tailscale.Hostname,
			AuthKey:   cfg.Tailscale.AuthKey,
			Routes:    cfg.Tailscale.Routes,
			SSH:       cfg.Tailscale.SSH,
			ExtraArgs: cfg.Tailscale.ExtraArgs,
		},
		SwanConfigPath:  cfg.Swan.ConfigPath,
		SwanAutoStart:   cfg.Swan.AutoStart,
		SwanConnections: cfg.Swan.Connections,
	}

	if err := supervisor.SetupSystem(); err != nil {
		slog.Error("System setup failed", "error", err)
		os.Exit(1)
	}

	sup := supervisor.New(supervisorCfg)

	if err := sup.Start(ctx); err != nil {
		slog.Error("Supervisor start failed", "error", err)
		os.Exit(1)
	}

	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", sig)
		sup.Stop()
	case err := <-sup.Errors():
		slog.Error("Process died", "error", err)
		sup.Stop()
		os.Exit(1)
	}
}
