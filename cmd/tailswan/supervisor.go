package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/supervisor"
)

func runSupervisor() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	cfg := loadConfig()

	if err := supervisor.SetupSystem(); err != nil {
		log.Fatalf("System setup failed: %v", err)
	}

	sup := supervisor.New(cfg)

	if err := sup.Start(ctx); err != nil {
		log.Fatalf("Supervisor start failed: %v", err)
	}

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %s, shutting down...", sig)
		sup.Stop()
	case err := <-sup.Errors():
		log.Printf("Process died: %v", err)
		sup.Stop()
		os.Exit(1)
	}
}

func loadConfig() supervisor.Config {
	cfg := config.Load()

	return supervisor.Config{
		ControlPort:       cfg.Port,
		TailscaleStateDir: cfg.Tailscale.StateDir,
		TailscaleSocket:   cfg.Tailscale.Socket,
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
}
