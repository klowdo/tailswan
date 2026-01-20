package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Config struct {
	ControlPort       string
	TailscaleStateDir string
	TailscaleSocket   string
	UseTsnet          bool
	TailscaleConfig   TailscaleConfig
	SwanConfigPath    string
	SwanAutoStart     bool
	SwanConnections   []string
}

type Supervisor struct {
	config Config

	ipsec      *Process
	tailscaled *Process
	server     *Process

	tsService   *TailscaleService
	swanService *SwanService

	errors chan error
}

func New(cfg Config) *Supervisor {
	return &Supervisor{
		config:      cfg,
		ipsec:       &Process{},
		tailscaled:  &Process{},
		server:      &Process{},
		tsService:   NewTailscaleService(),
		swanService: &SwanService{},
		errors:      make(chan error, 1),
	}
}

func (s *Supervisor) Start(ctx context.Context) error {
	slog.Info("Starting strongSwan charon daemon")
	if err := s.ipsec.Start("ipsec", "start", "--nofork"); err != nil {
		return fmt.Errorf("ipsec start: %w", err)
	}
	time.Sleep(2 * time.Second)

	if err := s.swanService.LoadConfig(s.config.SwanConfigPath); err != nil {
		slog.Warn("swanctl load failed", "error", err)
	}

	if s.config.SwanAutoStart {
		for _, conn := range s.config.SwanConnections {
			if err := s.swanService.Initiate(conn); err != nil {
				slog.Warn("Failed to start connection", "connection", conn, "error", err)
			}
		}
	}

	slog.Info("Starting control server", "port", s.config.ControlPort)
	if err := s.server.Start("controlserver"); err != nil {
		return fmt.Errorf("controlserver start: %w", err)
	}

	if s.config.UseTsnet {
		slog.Info("Using tsnet for Tailscale integration (embedded)")
		slog.Info("Control server will handle Tailscale connectivity via tsnet")
	} else {
		slog.Info("Starting tailscaled")
		if err := s.tailscaled.Start(
			"tailscaled",
			"--state", fmt.Sprintf("%s/tailscaled.state", s.config.TailscaleStateDir),
			"--socket", s.config.TailscaleSocket,
			"--tun", "userspace-networking",
		); err != nil {
			return fmt.Errorf("tailscaled start: %w", err)
		}

		slog.Info("Waiting for tailscaled to be ready")
		readyCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		if err := s.tsService.WaitReady(readyCtx); err != nil {
			return fmt.Errorf("tailscaled not ready: %w", err)
		}

		slog.Info("Bringing up Tailscale")
		if err := s.tsService.Up(s.config.TailscaleConfig); err != nil {
			return fmt.Errorf("tailscale up: %w", err)
		}

		slog.Info("Enabling Tailscale Serve")
		if err := s.tsService.EnableServe(s.config.ControlPort); err != nil {
			return fmt.Errorf("tailscale serve: %w", err)
		}
	}

	s.printStatus()

	go s.monitor(ctx)

	return nil
}

func (s *Supervisor) Stop() {
	slog.Info("Shutting down")

	if s.server != nil {
		s.server.Kill()
	}

	if s.tailscaled != nil {
		s.tailscaled.Kill()
	}

	if s.ipsec != nil {
		s.ipsec.Kill()
	}

	time.Sleep(2 * time.Second)
	slog.Info("Shutdown complete")
}

func (s *Supervisor) Errors() <-chan error {
	return s.errors
}

func (s *Supervisor) monitor(ctx context.Context) {
	errChan := make(chan error, 3)

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

	go func() {
		err := s.server.Wait()
		errChan <- fmt.Errorf("controlserver exited: %w", err)
	}()

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
