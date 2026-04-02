package supervisor

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"tailscale.com/client/local"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/server"
	"github.com/klowdo/tailswan/internal/swan"
)

type Config struct {
	ControlPort       string
	TailscaleStateDir string
	TailscaleSocket   string
	SwanConfigPath    string
	SwanConnections   []string
	TailscaleConfig   TailscaleConfig
	UseTsnet          bool
	SwanAutoStart     bool
}

func (c *Config) ServerConfig() *config.Config {
	return &config.Config{
		Port: c.ControlPort,
	}
}

type Supervisor struct {
	ipsec      *Process
	tailscaled *Process
	tsService  *TailscaleService
	swanSvc    *swan.Service
	srv        *server.Server
	errors     chan error
	webFS      embed.FS
	config     Config
}

func New(cfg *Config, webFS embed.FS) *Supervisor {
	return &Supervisor{
		config:     *cfg,
		ipsec:      &Process{},
		tailscaled: &Process{},
		tsService:  NewTailscaleService(),
		errors:     make(chan error, 1),
		webFS:      webFS,
	}
}

func (s *Supervisor) Start(ctx context.Context) error {
	if err := s.startIPsec(); err != nil {
		return err
	}

	s.initSwanService()

	if !s.config.UseTsnet {
		if err := s.startTailscale(ctx); err != nil {
			return err
		}
	}

	if err := s.startServer(); err != nil {
		return err
	}

	s.printStatus()
	go s.monitor(ctx)

	return nil
}

func (s *Supervisor) startIPsec() error {
	slog.Info("Starting strongSwan charon daemon")
	if err := s.ipsec.Start("ipsec", "start", "--nofork"); err != nil {
		return fmt.Errorf("ipsec start: %w", err)
	}
	time.Sleep(2 * time.Second)
	return nil
}

func (s *Supervisor) initSwanService() {
	svc, err := swan.NewService()
	if err != nil {
		slog.Warn("VICI connection failed", "error", err)
		return
	}

	s.swanSvc = svc
	if loadErr := s.swanSvc.LoadAll(); loadErr != nil {
		slog.Warn("swanctl load failed", "error", loadErr)
	}

	if s.config.SwanAutoStart {
		for _, conn := range s.config.SwanConnections {
			if initErr := s.swanSvc.Initiate(conn); initErr != nil {
				slog.Warn("Failed to start connection", "connection", conn, "error", initErr)
			}
		}
	}
}

func (s *Supervisor) startTailscale(ctx context.Context) error {
	slog.Info("Starting tailscaled",
		"state_dir", s.config.TailscaleStateDir,
		"socket", s.config.TailscaleSocket)
	if err := s.tailscaled.Start(
		"tailscaled",
		"--state", fmt.Sprintf("%s/tailscaled.state", s.config.TailscaleStateDir),
		"--socket", s.config.TailscaleSocket,
		"--tun", "userspace-networking",
	); err != nil {
		return fmt.Errorf("tailscaled start: %w", err)
	}
	slog.Info("tailscaled process started")

	slog.Info("Waiting for tailscaled to be ready")
	readyCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := s.tsService.WaitReady(readyCtx); err != nil {
		return fmt.Errorf("tailscaled not ready: %w", err)
	}

	slog.Info("Bringing up Tailscale")
	if err := s.tsService.Up(&s.config.TailscaleConfig); err != nil {
		return fmt.Errorf("tailscale up: %w", err)
	}

	if s.config.TailscaleConfig.EnableServe {
		slog.Info("Enabling Tailscale Serve", "port", s.config.ControlPort)
		if err := s.tsService.EnableServe(s.config.ControlPort); err != nil {
			return fmt.Errorf("tailscale serve: %w", err)
		}
	}

	return nil
}

func (s *Supervisor) startServer() error {
	tsClient := &local.Client{}

	srv, err := server.New(s.config.ServerConfig(), s.webFS, s.swanSvc, tsClient)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	s.srv = srv

	go func() {
		var serverErr error
		if s.config.UseTsnet {
			slog.Info("Starting server with tsnet")
			serverErr = s.srv.StartWithTsnet(
				s.config.TailscaleConfig.Hostname,
				s.config.TailscaleConfig.AuthKey,
				s.config.TailscaleConfig.Routes,
			)
		} else {
			serverErr = s.srv.Start()
		}
		if serverErr != nil {
			s.errors <- fmt.Errorf("server: %w", serverErr)
		}
	}()

	return nil
}

func (s *Supervisor) Stop() {
	slog.Info("Shutting down")

	if s.srv != nil {
		s.srv.Close()
	}

	if s.tailscaled != nil {
		if err := s.tailscaled.Kill(); err != nil {
			slog.Error("Failed to kill tailscaled", "error", err)
		}
	}

	if s.ipsec != nil {
		if err := s.ipsec.Kill(); err != nil {
			slog.Error("Failed to kill ipsec", "error", err)
		}
	}

	if s.swanSvc != nil {
		if err := s.swanSvc.Close(); err != nil {
			slog.Error("Failed to close VICI session", "error", err)
		}
	}

	time.Sleep(2 * time.Second)
	slog.Info("Shutdown complete")
}

func (s *Supervisor) Errors() <-chan error {
	return s.errors
}

func (s *Supervisor) monitor(ctx context.Context) {
	errChan := make(chan error, 2)

	go func() {
		err := s.ipsec.Wait()
		errChan <- fmt.Errorf("ipsec exited: %w", err)
	}()

	if !s.config.UseTsnet {
		go func() {
			err := s.tailscaled.Wait()
			errChan <- fmt.Errorf("tailscaled exited: %w", err)
		}()
	}

	select {
	case <-ctx.Done():
		return
	case err := <-errChan:
		s.errors <- err
	}
}

func (s *Supervisor) printStatus() {
	slog.Info("")
	slog.Info("===========================================")
	slog.Info("TailSwan is running")
	slog.Info("===========================================")
	slog.Info("Control server running", "url", fmt.Sprintf("http://localhost:%s", s.config.ControlPort))
	slog.Info("Access via Tailscale Serve (HTTP and HTTPS)")
	slog.Info("")
}
