package main

import (
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/server"
	"github.com/spf13/cobra"
)

//go:embed web
var webFS embed.FS

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
	Use:   "controlserver",
	Short: "TailSwan Control Server",
	Long:  `HTTP control server for managing TailSwan VPN connections.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		cfg = config.Load()
		initLogger(cfg)
	},
	Run: func(cmd *cobra.Command, args []string) {
		srv, err := server.New(cfg, webFS)
		if err != nil {
			slog.Error("Failed to create server", "error", err)
			os.Exit(1)
		}
		defer srv.Close()

		if cfg.Tailscale.UseTsnet {
			slog.Info("Starting server with tsnet")
			if err := srv.StartWithTsnet(
				cfg.Tailscale.Hostname,
				cfg.Tailscale.AuthKey,
				cfg.Tailscale.Routes,
			); err != nil {
				slog.Error("Server failed", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("Starting server with Tailscale Serve")
			if err := srv.Start(); err != nil {
				slog.Error("Server failed", "error", err)
				os.Exit(1)
			}
		}
	},
}

func initLogger(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}
