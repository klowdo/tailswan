package handlers

import (
	"net/http"

	"github.com/klowdo/tailswan/internal/models"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, models.Response{
		Success: true,
		Message: "TailSwan control server is healthy",
	})
}
