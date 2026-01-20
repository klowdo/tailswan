package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/klowdo/tailswan/internal/models"
	"github.com/klowdo/tailswan/internal/sse"
)

type SSEHandler struct {
	broadcaster *sse.EventBroadcaster
}

func NewSSEHandler(broadcaster *sse.EventBroadcaster) *SSEHandler {
	return &SSEHandler{
		broadcaster: broadcaster,
	}
}

func (h *SSEHandler) Events(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	clientChan := make(chan models.SSEMessage, 10)
	h.broadcaster.RegisterClient(clientChan)
	defer h.broadcaster.UnregisterClient(clientChan)

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "event: %s\n", msg.Event)
			fmt.Fprintf(w, "data: %s\n\n", string(msg.Data))
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
