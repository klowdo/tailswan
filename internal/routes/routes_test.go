package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klowdo/tailswan/internal/handlers"
)

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()

	viciHandler := &handlers.VICIHandler{}
	tsHandler := &handlers.TailscaleHandler{}
	healthHandler := &handlers.HealthHandler{}
	sseHandler := &handlers.SSEHandler{}

	RegisterRoutes(mux, viciHandler, tsHandler, healthHandler, sseHandler)

	endpoints := []string{
		"/api/health",
		"/api/events",
		"/api/vici/connections/up",
		"/api/vici/connections/down",
		"/api/vici/connections/list",
		"/api/vici/sas/list",
		"/api/tailscale/status",
		"/api/tailscale/peers",
		"/api/tailscale/serve",
		"/api/tailscale/whois",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			defer func() { recover() }() //nolint:errcheck

			req := httptest.NewRequest(http.MethodGet, endpoint, http.NoBody)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Errorf("route %s not registered", endpoint)
			}
		})
	}
}

func TestUnregisteredRouteReturns404(t *testing.T) {
	mux := http.NewServeMux()

	viciHandler := &handlers.VICIHandler{}
	tsHandler := &handlers.TailscaleHandler{}
	healthHandler := &handlers.HealthHandler{}
	sseHandler := &handlers.SSEHandler{}

	RegisterRoutes(mux, viciHandler, tsHandler, healthHandler, sseHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", http.NoBody)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unregistered route, got %d", rr.Code)
	}
}
