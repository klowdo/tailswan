package supervisor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
)

type TailscaleService struct {
	client *tailscale.LocalClient
}

func NewTailscaleService() *TailscaleService {
	return &TailscaleService{
		client: &tailscale.LocalClient{},
	}
}

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

			_, err := ts.client.Status(ctx)
			if err == nil {
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
	ctx := context.Background()

	config := &ipn.ServeConfig{
		TCP: map[uint16]*ipn.TCPPortHandler{
			443: {HTTPS: true},
			80:  {HTTP: true},
		},
		Web: map[ipn.HostPort]*ipn.WebServerConfig{
			"${TS_CERT_DOMAIN}:443": {
				Handlers: map[string]*ipn.HTTPHandler{
					"/": {Proxy: "http://127.0.0.1:" + port},
				},
			},
			"${TS_CERT_DOMAIN}:80": {
				Handlers: map[string]*ipn.HTTPHandler{
					"/": {Proxy: "http://127.0.0.1:" + port},
				},
			},
		},
	}

	if err := ts.client.SetServeConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to set serve config: %w", err)
	}

	log.Println("✓ Control server available via Tailscale Serve (HTTP and HTTPS)")

	serveStatus, err := ts.client.GetServeConfig(ctx)
	if err == nil && serveStatus != nil {
		log.Printf("Serve config: %+v", serveStatus)
	}

	return nil
}
