package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/klowdo/tailswan/internal/cli"
	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/supervisor"
	"github.com/klowdo/tailswan/internal/version"
)

var cfg *config.Config

func main() {
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Get()),
	); err != nil {
		os.Exit(1)
	}
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "tailswan",
	Short: "TailSwan - IPsec & Tailscale VPN Supervisor",
	Long:  `TailSwan is a unified supervisor for managing strongSwan IPsec and Tailscale VPN connections.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg = config.Load()
		initLogger(cfg)
	},
}

func initLogger(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the TailSwan supervisor",
	Long:  `Start the TailSwan supervisor to manage strongSwan and Tailscale services.`,
	Run: func(cmd *cobra.Command, args []string) {
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
				StateDir:    cfg.Tailscale.StateDir,
				Socket:      cfg.Tailscale.Socket,
				Hostname:    cfg.Tailscale.Hostname,
				AuthKey:     cfg.Tailscale.AuthKey,
				Routes:      cfg.Tailscale.Routes,
				SSH:         cfg.Tailscale.SSH,
				ExtraArgs:   cfg.Tailscale.ExtraArgs,
				EnableServe: cfg.Tailscale.EnableServe,
			},
			SwanConfigPath:  cfg.Swan.ConfigPath,
			SwanAutoStart:   cfg.Swan.AutoStart,
			SwanConnections: cfg.Swan.Connections,
		}

		if err := supervisor.SetupSystem(); err != nil {
			slog.Error("System setup failed", "error", err)
			os.Exit(1)
		}

		sup := supervisor.New(&supervisorCfg)

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
	},
}

func init() {
	rootCmd.AddCommand(
		serveCmd,
		cli.NewHealthCheckCmd(),
		cli.NewStatusCmd(),
		cli.NewConnectionsCmd(),
		cli.NewSAsCmd(),
		cli.NewStartCmd(),
		cli.NewStopCmd(),
		cli.NewReloadCmd(),
	)
}
