package sse

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"tailscale.com/client/local"

	"github.com/klowdo/tailswan/internal/models"
	"github.com/klowdo/tailswan/internal/swan"
)

type poller struct {
	Fetch    func() (any, error)
	Name     string
	Event    string
	Interval time.Duration
}

type EventBroadcaster struct {
	ctx             context.Context
	clients         map[chan models.SSEMessage]bool
	swanSvc         *swan.Service
	tailscaleClient *local.Client
	stateTracker    *StateTracker
	cancel          context.CancelFunc
	clientsMux      sync.RWMutex
}

func NewEventBroadcaster(swanSvc *swan.Service, tsClient *local.Client) *EventBroadcaster {
	return &EventBroadcaster{
		clients:         make(map[chan models.SSEMessage]bool),
		swanSvc:         swanSvc,
		tailscaleClient: tsClient,
		stateTracker:    NewStateTracker(),
	}
}

func (eb *EventBroadcaster) SetTailscaleClient(client *local.Client) {
	eb.tailscaleClient = client
}

func (eb *EventBroadcaster) Start(ctx context.Context) {
	eb.ctx, eb.cancel = context.WithCancel(ctx)

	pollers := []poller{
		{Name: "sas", Event: "sa-update", Interval: 5 * time.Second, Fetch: eb.fetchSAs},
		{Name: "peers", Event: "peer-update", Interval: 10 * time.Second, Fetch: eb.fetchPeers},
		{Name: "connections", Event: "connection-update", Interval: 30 * time.Second, Fetch: eb.fetchConnections},
		{Name: "node", Event: "node-update", Interval: 30 * time.Second, Fetch: eb.fetchNodeStatus},
	}

	for _, p := range pollers {
		go eb.poll(eb.ctx, p)
	}

	<-eb.ctx.Done()
}

func (eb *EventBroadcaster) Stop() {
	if eb.cancel != nil {
		eb.cancel()
	}

	eb.clientsMux.Lock()
	defer eb.clientsMux.Unlock()

	for clientChan := range eb.clients {
		close(clientChan)
	}
	eb.clients = make(map[chan models.SSEMessage]bool)
}

func (eb *EventBroadcaster) RegisterClient(ch chan models.SSEMessage) {
	eb.clientsMux.Lock()
	defer eb.clientsMux.Unlock()
	eb.clients[ch] = true
	slog.Info("SSE client connected", "total", len(eb.clients))
}

func (eb *EventBroadcaster) UnregisterClient(ch chan models.SSEMessage) {
	eb.clientsMux.Lock()
	defer eb.clientsMux.Unlock()
	if _, ok := eb.clients[ch]; ok {
		delete(eb.clients, ch)
		close(ch)
		slog.Info("SSE client disconnected", "total", len(eb.clients))
	}
}

func (eb *EventBroadcaster) broadcast(msg models.SSEMessage) {
	eb.clientsMux.RLock()
	defer eb.clientsMux.RUnlock()

	for clientChan := range eb.clients {
		select {
		case clientChan <- msg:
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (eb *EventBroadcaster) poll(ctx context.Context, p poller) {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := p.Fetch()
			if err != nil {
				continue
			}
			if eb.stateTracker.HasChanged(p.Name, result) {
				data, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					continue
				}
				eb.broadcast(models.SSEMessage{Event: p.Event, Data: data})
			}
		}
	}
}

func (eb *EventBroadcaster) fetchSAs() (any, error) {
	sas, err := eb.swanSvc.ListSAs()
	if err != nil {
		slog.Info("Error fetching SAs", "error", err)
		return map[string]any{"success": false, "sas": []map[string]any{}}, nil
	}
	return map[string]any{"success": true, "sas": sas}, nil
}

func (eb *EventBroadcaster) fetchPeers() (any, error) {
	ctx := context.Background()
	status, err := eb.tailscaleClient.Status(ctx)
	if err != nil {
		slog.Info("Error fetching peers", "error", err)
		return map[string]any{"success": false, "peers": []map[string]any{}}, nil
	}

	var peers []map[string]any
	for _, peer := range status.Peer {
		peerInfo := map[string]any{
			"id":            peer.ID,
			"hostname":      peer.HostName,
			"dns_name":      peer.DNSName,
			"online":        peer.Online,
			"tailscale_ips": peer.TailscaleIPs,
			"os":            peer.OS,
			"last_seen":     peer.LastSeen,
		}
		peers = append(peers, peerInfo)
	}

	return map[string]any{
		"success": true,
		"peers":   peers,
		"self":    status.Self,
	}, nil
}

func (eb *EventBroadcaster) fetchConnections() (any, error) {
	connections, err := eb.swanSvc.ListConnections()
	if err != nil {
		slog.Info("Error fetching connections", "error", err)
		return map[string]any{"success": false, "connections": []map[string]any{}}, nil
	}
	return map[string]any{"success": true, "connections": connections}, nil
}

func (eb *EventBroadcaster) fetchNodeStatus() (any, error) {
	ctx := context.Background()
	status, err := eb.tailscaleClient.Status(ctx)
	if err != nil {
		slog.Info("Error fetching node status", "error", err)
		return map[string]any{"success": false}, nil
	}

	return map[string]any{
		"success": true,
		"status": map[string]any{
			"BackendState": status.BackendState,
			"Self":         status.Self,
		},
	}, nil
}
