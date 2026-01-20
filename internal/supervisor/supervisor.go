package supervisor

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Config struct {
	ControlPort       string
	TailscaleStateDir string
	TailscaleSocket   string
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
		tsService:   &TailscaleService{},
		swanService: &SwanService{},
		errors:      make(chan error, 1),
	}
}

func (s *Supervisor) Start(ctx context.Context) error {
	log.Println("Starting strongSwan charon daemon...")
	if err := s.ipsec.Start("ipsec", "start", "--nofork"); err != nil {
		return fmt.Errorf("ipsec start: %w", err)
	}
	time.Sleep(2 * time.Second)

	if err := s.swanService.LoadConfig(s.config.SwanConfigPath); err != nil {
		log.Printf("Warning: swanctl load: %v", err)
	}

	if s.config.SwanAutoStart {
		for _, conn := range s.config.SwanConnections {
			if err := s.swanService.Initiate(conn); err != nil {
				log.Printf("Warning: Failed to start %s: %v", conn, err)
			}
		}
	}

	log.Printf("Starting control server on port %s...", s.config.ControlPort)
	if err := s.server.Start("controlserver"); err != nil {
		return fmt.Errorf("controlserver start: %w", err)
	}

	log.Println("Starting tailscaled...")
	if err := s.tailscaled.Start(
		"tailscaled",
		"--state", fmt.Sprintf("%s/tailscaled.state", s.config.TailscaleStateDir),
		"--socket", s.config.TailscaleSocket,
		"--tun", "userspace-networking",
	); err != nil {
		return fmt.Errorf("tailscaled start: %w", err)
	}

	log.Println("Waiting for tailscaled to be ready...")
	readyCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := s.tsService.WaitReady(readyCtx); err != nil {
		return fmt.Errorf("tailscaled not ready: %w", err)
	}

	log.Println("Bringing up Tailscale...")
	if err := s.tsService.Up(s.config.TailscaleConfig); err != nil {
		return fmt.Errorf("tailscale up: %w", err)
	}

	log.Println("Enabling Tailscale Serve...")
	if err := s.tsService.EnableServe(s.config.ControlPort); err != nil {
		return fmt.Errorf("tailscale serve: %w", err)
	}

	s.printStatus()

	go s.monitor(ctx)

	return nil
}

func (s *Supervisor) Stop() {
	log.Println("Shutting down...")

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
	log.Println("Shutdown complete")
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

	go func() {
		err := s.tailscaled.Wait()
		errChan <- fmt.Errorf("tailscaled exited: %w", err)
	}()

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
	log.Println("")
	log.Println("===========================================")
	log.Println("TailSwan is running")
	log.Println("===========================================")
	log.Printf("Control server: http://localhost:%s", s.config.ControlPort)
	log.Println("Access via Tailscale Serve (HTTP and HTTPS)")
	log.Println("")
}
