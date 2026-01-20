package sse

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/klowdo/tailswan/internal/models"
	"github.com/strongswan/govici/vici"
	"tailscale.com/client/tailscale"
)

type EventBroadcaster struct {
	clients    map[chan models.SSEMessage]bool
	clientsMux sync.RWMutex

	viciSession      *vici.Session
	tailscaleClient  *tailscale.LocalClient
	stateTracker     *StateTracker

	ctx    context.Context
	cancel context.CancelFunc
}

func NewEventBroadcaster(viciSession *vici.Session, tsClient *tailscale.LocalClient) *EventBroadcaster {
	return &EventBroadcaster{
		clients:         make(map[chan models.SSEMessage]bool),
		viciSession:     viciSession,
		tailscaleClient: tsClient,
		stateTracker:    NewStateTracker(),
	}
}

func (eb *EventBroadcaster) Start(ctx context.Context) {
	eb.ctx, eb.cancel = context.WithCancel(ctx)

	go eb.pollSAs(eb.ctx)
	go eb.pollPeers(eb.ctx)
	go eb.pollConnections(eb.ctx)
	go eb.pollNodeStatus(eb.ctx)

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
	log.Printf("SSE client connected (total: %d)", len(eb.clients))
}

func (eb *EventBroadcaster) UnregisterClient(ch chan models.SSEMessage) {
	eb.clientsMux.Lock()
	defer eb.clientsMux.Unlock()
	if _, ok := eb.clients[ch]; ok {
		delete(eb.clients, ch)
		close(ch)
		log.Printf("SSE client disconnected (total: %d)", len(eb.clients))
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

func (eb *EventBroadcaster) pollSAs(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sas := eb.fetchSAs()
			if eb.stateTracker.HasChanged("sas", sas) {
				data, err := json.Marshal(sas)
				if err == nil {
					eb.broadcast(models.SSEMessage{
						Event: "sa-update",
						Data:  data,
					})
				}
			}
		}
	}
}

func (eb *EventBroadcaster) pollPeers(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers := eb.fetchPeers()
			if eb.stateTracker.HasChanged("peers", peers) {
				data, err := json.Marshal(peers)
				if err == nil {
					eb.broadcast(models.SSEMessage{
						Event: "peer-update",
						Data:  data,
					})
				}
			}
		}
	}
}

func (eb *EventBroadcaster) pollConnections(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			connections := eb.fetchConnections()
			if eb.stateTracker.HasChanged("connections", connections) {
				data, err := json.Marshal(connections)
				if err == nil {
					eb.broadcast(models.SSEMessage{
						Event: "connection-update",
						Data:  data,
					})
				}
			}
		}
	}
}

func (eb *EventBroadcaster) pollNodeStatus(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status := eb.fetchNodeStatus()
			if eb.stateTracker.HasChanged("node", status) {
				data, err := json.Marshal(status)
				if err == nil {
					eb.broadcast(models.SSEMessage{
						Event: "node-update",
						Data:  data,
					})
				}
			}
		}
	}
}

func (eb *EventBroadcaster) fetchSAs() map[string]interface{} {
	msg := vici.NewMessage()
	messages, err := eb.viciSession.StreamedCommandRequest("list-sas", "list-sa", msg)
	if err != nil {
		log.Printf("Error fetching SAs: %v", err)
		return map[string]interface{}{"success": false, "sas": []map[string]interface{}{}}
	}

	var sas []map[string]interface{}
	for _, m := range messages.Messages() {
		saMap := make(map[string]interface{})
		for _, key := range m.Keys() {
			saMap[key] = m.Get(key)
		}
		sas = append(sas, saMap)
	}

	return map[string]interface{}{
		"success": true,
		"sas":     sas,
	}
}

func (eb *EventBroadcaster) fetchPeers() map[string]interface{} {
	ctx := context.Background()
	status, err := eb.tailscaleClient.Status(ctx)
	if err != nil {
		log.Printf("Error fetching peers: %v", err)
		return map[string]interface{}{"success": false, "peers": []map[string]interface{}{}}
	}

	var peers []map[string]interface{}
	for _, peer := range status.Peer {
		peerInfo := map[string]interface{}{
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

	return map[string]interface{}{
		"success": true,
		"peers":   peers,
		"self":    status.Self,
	}
}

func (eb *EventBroadcaster) fetchConnections() map[string]interface{} {
	msg := vici.NewMessage()
	messages, err := eb.viciSession.StreamedCommandRequest("list-conns", "list-conn", msg)
	if err != nil {
		log.Printf("Error fetching connections: %v", err)
		return map[string]interface{}{"success": false, "connections": []map[string]interface{}{}}
	}

	var connections []map[string]interface{}
	for _, m := range messages.Messages() {
		connMap := make(map[string]interface{})
		for _, key := range m.Keys() {
			connMap[key] = m.Get(key)
		}
		connections = append(connections, connMap)
	}

	return map[string]interface{}{
		"success":     true,
		"connections": connections,
	}
}

func (eb *EventBroadcaster) fetchNodeStatus() map[string]interface{} {
	ctx := context.Background()
	status, err := eb.tailscaleClient.Status(ctx)
	if err != nil {
		log.Printf("Error fetching node status: %v", err)
		return map[string]interface{}{"success": false}
	}

	return map[string]interface{}{
		"success": true,
		"status": map[string]interface{}{
			"BackendState": status.BackendState,
			"Self":         status.Self,
		},
	}
}
