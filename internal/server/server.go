package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/handlers"
	"github.com/klowdo/tailswan/internal/routes"
	"github.com/klowdo/tailswan/internal/sse"
)

type Server struct {
	config        *config.Config
	viciHandler   *handlers.VICIHandler
	tsHandler     *handlers.TailscaleHandler
	healthHandler *handlers.HealthHandler
	broadcaster   *sse.EventBroadcaster
	cancel        context.CancelFunc
	mux           *http.ServeMux
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

func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.broadcaster != nil {
		s.broadcaster.Stop()
	}

	if s.viciHandler != nil {
		s.viciHandler.Close()
	}
}
