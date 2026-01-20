package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
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
	log.Printf("Starting TailSwan control server on %s", addr)
	log.Println("Web UI available at: http://localhost:%s/", s.config.Port)
	log.Println("")
	log.Println("API endpoints:")
	log.Println("  GET  /api/health                      - Health check")
	log.Println("  GET  /api/events                      - Server-Sent Events stream")
	log.Println("")
	log.Println("  VICI (strongSwan):")
	log.Println("    POST /api/vici/connections/up       - Bring connection up")
	log.Println("    POST /api/vici/connections/down     - Bring connection down")
	log.Println("    GET  /api/vici/connections/list     - List all connections")
	log.Println("    GET  /api/vici/sas/list             - List security associations")
	log.Println("")
	log.Println("  Tailscale:")
	log.Println("    GET  /api/tailscale/status          - Tailscale status")
	log.Println("    GET  /api/tailscale/peers           - List all peers")
	log.Println("    GET  /api/tailscale/serve           - Tailscale Serve configuration")
	log.Println("    GET  /api/tailscale/whois           - WhoIs lookup")

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

	log.Println("Waiting for tsnet to be ready...")
	dnsName := ""
	for i := 0; i < 30; i++ {
		st, e := localClient.StatusWithoutPeers(ctx)
		if e == nil && st.Self.DNSName != "" {
			dnsName = st.Self.DNSName
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Printf("Starting TailSwan control server")
	log.Printf("Local access: http://localhost:%s/", s.config.Port)
	if dnsName != "" {
		dnsName = strings.TrimSuffix(dnsName, ".")
		log.Printf("Tailscale access: https://%s/", dnsName)
	} else {
		log.Println("Tailscale DNS name not yet available")
	}
	log.Println("")
	log.Println("API endpoints:")
	log.Println("  GET  /api/health                      - Health check")
	log.Println("  GET  /api/events                      - Server-Sent Events stream")
	log.Println("")
	log.Println("  VICI (strongSwan):")
	log.Println("    POST /api/vici/connections/up       - Bring connection up")
	log.Println("    POST /api/vici/connections/down     - Bring connection down")
	log.Println("    GET  /api/vici/connections/list     - List all connections")
	log.Println("    GET  /api/vici/sas/list             - List security associations")
	log.Println("")
	log.Println("  Tailscale:")
	log.Println("    GET  /api/tailscale/status          - Tailscale status")
	log.Println("    GET  /api/tailscale/peers           - List all peers")
	log.Println("    GET  /api/tailscale/serve           - Tailscale Serve configuration")
	log.Println("    GET  /api/tailscale/whois           - WhoIs lookup")

	go func() {
		log.Println("Starting tsnet HTTPS server on :443...")
		if err := http.Serve(s.tsnetListener, s.mux); err != nil {
			log.Printf("tsnet server error: %v", err)
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
