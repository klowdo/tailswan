package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/handlers"
	"github.com/klowdo/tailswan/internal/routes"
	"github.com/klowdo/tailswan/internal/sse"
	"tailscale.com/tsnet"
)

type Server struct {
	config        *config.Config
	viciHandler   *handlers.VICIHandler
	tsHandler     *handlers.TailscaleHandler
	healthHandler *handlers.HealthHandler
	broadcaster   *sse.EventBroadcaster
	cancel        context.CancelFunc
	mux           *http.ServeMux
	tsnetServer   *tsnet.Server
	tsnetListener net.Listener
}

func New(cfg *config.Config, webFS embed.FS) (*Server, error) {
	viciHandler, err := handlers.NewVICIHandler()
	if err != nil {
		return nil, err
	}

	tsHandler := handlers.NewTailscaleHandler()
	healthHandler := handlers.NewHealthHandler()

	broadcaster := sse.NewEventBroadcaster(viciHandler.Session(), tsHandler.LocalClient())
	sseHandler := handlers.NewSSEHandler(broadcaster)

	mux := http.NewServeMux()

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, err
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(webContent))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			content, err := webFS.ReadFile("web/index.html")
			if err != nil {
				http.Error(w, "Failed to load page", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(content)
		} else {
			http.NotFound(w, r)
		}
	})

	routes.RegisterRoutes(mux, viciHandler, tsHandler, healthHandler, sseHandler)

	return &Server{
		config:        cfg,
		viciHandler:   viciHandler,
		tsHandler:     tsHandler,
		healthHandler: healthHandler,
		broadcaster:   broadcaster,
		mux:           mux,
	}, nil
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.broadcaster.Start(ctx)

	addr := s.config.Address()
	slog.Info("Starting TailSwan control server", "address", addr)
	slog.Info("Web UI available", "url", fmt.Sprintf("http://localhost:%s/", s.config.Port))
	slog.Info("")
	slog.Info("API endpoints:")
	slog.Info("  GET  /api/health                      - Health check")
	slog.Info("  GET  /api/events                      - Server-Sent Events stream")
	slog.Info("")
	slog.Info("  VICI (strongSwan):")
	slog.Info("    POST /api/vici/connections/up       - Bring connection up")
	slog.Info("    POST /api/vici/connections/down     - Bring connection down")
	slog.Info("    GET  /api/vici/connections/list     - List all connections")
	slog.Info("    GET  /api/vici/sas/list             - List security associations")
	slog.Info("")
	slog.Info("  Tailscale:")
	slog.Info("    GET  /api/tailscale/status          - Tailscale status")
	slog.Info("    GET  /api/tailscale/peers           - List all peers")
	slog.Info("    GET  /api/tailscale/serve           - Tailscale Serve configuration")
	slog.Info("    GET  /api/tailscale/whois           - WhoIs lookup")

	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) StartWithTsnet(hostname, authKey string, routes []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.broadcaster.Start(ctx)

	s.tsnetServer = &tsnet.Server{
		Hostname:  hostname,
		AuthKey:   authKey,
		Dir:       "/var/lib/tailscale",
		Ephemeral: false,
	}

	var err error
	s.tsnetListener, err = s.tsnetServer.ListenTLS("tcp", ":443")
	if err != nil {
		return err
	}

	localClient, err := s.tsnetServer.LocalClient()
	if err != nil {
		return err
	}

	s.tsHandler.SetClient(localClient)
	s.broadcaster.SetTailscaleClient(localClient)

	slog.Info("Waiting for tsnet to be ready...")
	dnsName := ""
	for i := 0; i < 30; i++ {
		st, e := localClient.StatusWithoutPeers(ctx)
		if e == nil && st.Self.DNSName != "" {
			dnsName = st.Self.DNSName
			break
		}
		time.Sleep(1 * time.Second)
	}

	slog.Info("Starting TailSwan control server")
	slog.Info("Local access", "url", fmt.Sprintf("http://localhost:%s/", s.config.Port))
	if dnsName != "" {
		dnsName = strings.TrimSuffix(dnsName, ".")
		slog.Info("Tailscale access", "url", fmt.Sprintf("https://%s/", dnsName))
	} else {
		slog.Info("Tailscale DNS name not yet available")
	}
	slog.Info("")
	slog.Info("API endpoints:")
	slog.Info("  GET  /api/health                      - Health check")
	slog.Info("  GET  /api/events                      - Server-Sent Events stream")
	slog.Info("")
	slog.Info("  VICI (strongSwan):")
	slog.Info("    POST /api/vici/connections/up       - Bring connection up")
	slog.Info("    POST /api/vici/connections/down     - Bring connection down")
	slog.Info("    GET  /api/vici/connections/list     - List all connections")
	slog.Info("    GET  /api/vici/sas/list             - List security associations")
	slog.Info("")
	slog.Info("  Tailscale:")
	slog.Info("    GET  /api/tailscale/status          - Tailscale status")
	slog.Info("    GET  /api/tailscale/peers           - List all peers")
	slog.Info("    GET  /api/tailscale/serve           - Tailscale Serve configuration")
	slog.Info("    GET  /api/tailscale/whois           - WhoIs lookup")

	go func() {
		slog.Info("Starting tsnet HTTPS server on :443...")
		if err := http.Serve(s.tsnetListener, s.mux); err != nil {
			slog.Info("tsnet server error: %v", err)
		}
	}()

	addr := s.config.Address()
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.tsnetListener != nil {
		s.tsnetListener.Close()
	}

	if s.tsnetServer != nil {
		s.tsnetServer.Close()
	}

	if s.broadcaster != nil {
		s.broadcaster.Stop()
	}

	if s.viciHandler != nil {
		s.viciHandler.Close()
	}
}
