package supervisor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TailscaleService struct{}

type TailscaleConfig struct {
	StateDir  string
	Socket    string
	Hostname  string
	AuthKey   string
	Routes    []string
	SSH       bool
	ExtraArgs []string
}

func (ts *TailscaleService) WaitReady(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	attempts := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for tailscaled")
		case <-ticker.C:
			attempts++

			cmd := exec.Command("tailscale", "status")
			output, err := cmd.CombinedOutput()

			if err == nil || strings.Contains(string(output), "Logged out") {
				log.Println("✓ Tailscaled is ready")
				return nil
			}

			if attempts%10 == 0 {
				log.Printf("Still waiting for tailscaled... (%d/60 seconds)", attempts)
			}
		}
	}
}

func (ts *TailscaleService) Up(cfg TailscaleConfig) error {
	args := []string{"up", "--hostname=" + cfg.Hostname}

	if cfg.AuthKey != "" {
		args = append(args, "--authkey="+cfg.AuthKey)
	}

	if len(cfg.Routes) > 0 {
		routes := strings.Join(cfg.Routes, ",")
		args = append(args, "--advertise-routes="+routes)
		log.Printf("Advertising routes: %s", routes)
	}

	if cfg.SSH {
		args = append(args, "--ssh")
		log.Println("Enabling Tailscale SSH")
	}

	args = append(args, cfg.ExtraArgs...)

	log.Printf("Bringing up Tailscale: tailscale %s", strings.Join(args, " "))

	cmd := exec.Command("tailscale", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (ts *TailscaleService) EnableServe(port string) error {
	cmd := exec.Command("tailscale", "serve", "--bg", port)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("enable https serve: %w", err)
	}

	cmd = exec.Command("tailscale", "serve", "--bg", "--http", "80",
		"http://127.0.0.1:"+port)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("enable http serve: %w", err)
	}

	log.Println("✓ Control server available via Tailscale Serve (HTTP and HTTPS)")

	cmd = exec.Command("tailscale", "serve", "status")
	cmd.Stdout = os.Stdout
	cmd.Run()

	return nil
}
