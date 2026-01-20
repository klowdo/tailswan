package routes

import (
	"net/http"

	"github.com/klowdo/tailswan/internal/handlers"
)

func RegisterRoutes(mux *http.ServeMux, viciHandler *handlers.VICIHandler, tsHandler *handlers.TailscaleHandler, healthHandler *handlers.HealthHandler) {
	mux.HandleFunc("/api/health", healthHandler.Check)

	mux.HandleFunc("/api/vici/connections/up", viciHandler.ConnectionUp)
	mux.HandleFunc("/api/vici/connections/down", viciHandler.ConnectionDown)
	mux.HandleFunc("/api/vici/connections/list", viciHandler.ListConnections)
	mux.HandleFunc("/api/vici/sas/list", viciHandler.ListSAs)

	mux.HandleFunc("/api/tailscale/status", tsHandler.Status)
	mux.HandleFunc("/api/tailscale/peers", tsHandler.Peers)
	mux.HandleFunc("/api/tailscale/serve", tsHandler.ServeStatus)
	mux.HandleFunc("/api/tailscale/whois", tsHandler.WhoIs)
}
