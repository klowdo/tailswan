package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/strongswan/govici/vici"

	"github.com/klowdo/tailswan/internal/models"
)

type VICIHandler struct {
	session *vici.Session
}

func NewVICIHandler() (*VICIHandler, error) {
	session, err := vici.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create VICI session: %w", err)
	}

	return &VICIHandler{
		session: session,
	}, nil
}

func (h *VICIHandler) Close() error {
	if h.session != nil {
		return h.session.Close()
	}
	return nil
}

func (h *VICIHandler) Session() *vici.Session {
	return h.session
}

func (h *VICIHandler) ConnectionUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Connection name is required",
			Error:   "name field cannot be empty",
		})
		return
	}

	msg := vici.NewMessage()
	if err := msg.Set("child", req.Name); err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to set message field",
			Error:   err.Error(),
		})
		return
	}

	_, err := h.session.CommandRequest("initiate", msg)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to initiate connection '%s'", req.Name),
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, models.Response{
		Success: true,
		Message: fmt.Sprintf("Connection '%s' initiated successfully", req.Name),
	})
}

func (h *VICIHandler) ConnectionDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Connection name is required",
			Error:   "name field cannot be empty",
		})
		return
	}

	msg := vici.NewMessage()
	if err := msg.Set("child", req.Name); err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to set message field",
			Error:   err.Error(),
		})
		return
	}

	_, err := h.session.CommandRequest("terminate", msg)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to terminate connection '%s'", req.Name),
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, models.Response{
		Success: true,
		Message: fmt.Sprintf("Connection '%s' terminated successfully", req.Name),
	})
}

func (h *VICIHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	msg := vici.NewMessage()
	messages, err := h.session.StreamedCommandRequest("list-conns", "list-conn", msg)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to list connections",
			Error:   err.Error(),
		})
		return
	}

	connections := make([]map[string]interface{}, 0)
	for _, m := range messages.Messages() {
		connMap := make(map[string]interface{})
		for _, key := range m.Keys() {
			connMap[key] = m.Get(key)
		}
		connections = append(connections, connMap)
	}

	respondJSON(w, http.StatusOK, models.ConnectionsResponse{
		Success:     true,
		Connections: connections,
	})
}

func (h *VICIHandler) ListSAs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	msg := vici.NewMessage()
	messages, err := h.session.StreamedCommandRequest("list-sas", "list-sa", msg)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to list security associations",
			Error:   err.Error(),
		})
		return
	}

	sas := make([]map[string]interface{}, 0)
	for _, m := range messages.Messages() {
		saMap := make(map[string]interface{})
		for _, key := range m.Keys() {
			saMap[key] = m.Get(key)
		}
		sas = append(sas, saMap)
	}

	respondJSON(w, http.StatusOK, models.SAsResponse{
		Success: true,
		SAs:     sas,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
