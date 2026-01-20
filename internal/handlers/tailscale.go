package handlers

import (
	"context"
	"net/http"
	"sort"

	"github.com/klowdo/tailswan/internal/models"
	"tailscale.com/client/tailscale"
)

type TailscaleHandler struct {
	client *tailscale.LocalClient
}

func NewTailscaleHandler() *TailscaleHandler {
	return &TailscaleHandler{
		client: &tailscale.LocalClient{},
	}
}

func NewTailscaleHandlerWithClient(client *tailscale.LocalClient) *TailscaleHandler {
	return &TailscaleHandler{
		client: client,
	}
}

func (h *TailscaleHandler) LocalClient() *tailscale.LocalClient {
	return h.client
}

func (h *TailscaleHandler) SetClient(client *tailscale.LocalClient) {
	h.client = client
}

func (h *TailscaleHandler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	status, err := h.client.Status(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to get Tailscale status",
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  status,
	})
}

func (h *TailscaleHandler) Peers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	status, err := h.client.Status(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to get peers",
			Error:   err.Error(),
		})
		return
	}

	peers := make([]map[string]interface{}, 0)
	for _, peer := range status.Peer {
		peerInfo := map[string]interface{}{
			"id":            peer.ID,
			"hostname":      peer.HostName,
			"dns_name":      peer.DNSName,
			"online":        peer.Online,
			"tailscale_ips": peer.TailscaleIPs,
			"os":            peer.OS,
			"user_id":       peer.UserID,
			"exit_node":     peer.ExitNode,
			"last_seen":     peer.LastSeen,
			"tx_bytes":      peer.TxBytes,
			"rx_bytes":      peer.RxBytes,
			"created":       peer.Created,
			"tags":          peer.Tags,
		}
		peers = append(peers, peerInfo)
	}

	sort.Slice(peers, func(i, j int) bool {
		hostnameI, _ := peers[i]["hostname"].(string)
		hostnameJ, _ := peers[j]["hostname"].(string)
		return hostnameI < hostnameJ
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"self":    status.Self,
		"peers":   peers,
	})
}

func (h *TailscaleHandler) ServeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	serveConfig, err := h.client.GetServeConfig(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to get serve configuration",
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"config":  serveConfig,
	})
}

func (h *TailscaleHandler) WhoIs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	remoteAddr := r.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}

	ctx := context.Background()
	whois, err := h.client.WhoIs(ctx, remoteAddr)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to get WhoIs information",
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"whois":   whois,
	})
}
