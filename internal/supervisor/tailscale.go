package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	"tailscale.com/client/local"
	"tailscale.com/ipn"
)

type TailscaleService struct {
	client *local.Client
}

func NewTailscaleService() *TailscaleService {
	return &TailscaleService{
		client: &local.Client{},
	}
}

type TailscaleConfig struct {
	Hostname    string
	AuthKey     string `json:"-"`
	Routes      []string
	SSH         bool
	EnableServe bool
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
				slog.Info("Tailscaled is ready")
				return nil
			}

			if attempts%10 == 0 {
				slog.Info("Still waiting for tailscaled...", "attempts", attempts, "max", 60)
			}
		}
	}
}

func (ts *TailscaleService) Up(cfg *TailscaleConfig) error {
	ctx := context.Background()

	routes, err := parseRoutes(cfg.Routes)
	if err != nil {
		return fmt.Errorf("parse routes: %w", err)
	}

	prefs := &ipn.Prefs{
		WantRunning:     true,
		Hostname:        cfg.Hostname,
		RouteAll:        true,
		CorpDNS:         false,
		AdvertiseRoutes: routes,
		RunSSH:          cfg.SSH,
	}

	if len(routes) > 0 {
		slog.Info("Advertising routes", "routes", strings.Join(cfg.Routes, ","))
	}
	if cfg.SSH {
		slog.Info("Enabling Tailscale SSH")
	}

	opts := ipn.Options{
		UpdatePrefs: prefs,
		AuthKey:     cfg.AuthKey,
	}

	slog.Info("Bringing up Tailscale", "hostname", cfg.Hostname)

	if err := ts.client.Start(ctx, opts); err != nil {
		return fmt.Errorf("tailscale start failed: %w", err)
	}

	return nil
}

func (ts *TailscaleService) EnableServe(port string) error {
	ctx := context.Background()

	status, err := ts.client.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tailscale status: %w", err)
	}

	if status.Self == nil || status.Self.DNSName == "" {
		return fmt.Errorf("tailscale DNS name not available")
	}

	hostname := strings.TrimSuffix(status.Self.DNSName, ".")
	slog.Info("Configuring Tailscale Serve", "hostname", hostname, "port", port)

	config := &ipn.ServeConfig{
		TCP: map[uint16]*ipn.TCPPortHandler{
			443: {HTTPS: true},
			80:  {HTTP: true},
		},
		Web: map[ipn.HostPort]*ipn.WebServerConfig{
			ipn.HostPort(hostname + ":443"): {
				Handlers: map[string]*ipn.HTTPHandler{
					"/": {Proxy: "http://127.0.0.1:" + port},
				},
			},
			ipn.HostPort(hostname + ":80"): {
				Handlers: map[string]*ipn.HTTPHandler{
					"/": {Proxy: "http://127.0.0.1:" + port},
				},
			},
		},
	}

	if setErr := ts.client.SetServeConfig(ctx, config); setErr != nil {
		return fmt.Errorf("failed to set serve config: %w", setErr)
	}

	slog.Info("Control server available via Tailscale Serve", "url", "https://"+hostname)

	serveStatus, err := ts.client.GetServeConfig(ctx)
	if err == nil && serveStatus != nil {
		slog.Info("Serve config loaded", "config", serveStatus)
	}

	return nil
}

func parseRoutes(routes []string) ([]netip.Prefix, error) {
	if len(routes) == 0 {
		return nil, nil
	}

	prefixes := make([]netip.Prefix, 0, len(routes))
	for _, r := range routes {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(r)
		if err != nil {
			return nil, fmt.Errorf("invalid route %q: %w", r, err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}
