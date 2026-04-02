package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/klowdo/tailswan/internal/models"
	"github.com/klowdo/tailswan/internal/swan"
)

type VICIHandler struct {
	svc *swan.Service
}

func NewVICIHandler(svc *swan.Service) *VICIHandler {
	return &VICIHandler{svc: svc}
}

func (h *VICIHandler) SwanService() *swan.Service {
	return h.svc
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

	if err := h.svc.Initiate(req.Name); err != nil {
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

	if err := h.svc.Terminate(req.Name); err != nil {
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

	connections, err := h.svc.ListConnections()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to list connections",
			Error:   err.Error(),
		})
		return
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

	sas, err := h.svc.ListSAs()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to list security associations",
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, models.SAsResponse{
		Success: true,
		SAs:     sas,
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
