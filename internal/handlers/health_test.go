package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klowdo/tailswan/internal/models"
)

func TestHealthHandler_Check(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedBody   models.Response
		expectedStatus int
	}{
		{
			name:           "GET request returns healthy status",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody: models.Response{
				Success: true,
				Message: "TailSwan control server is healthy",
			},
		},
		{
			name:           "POST request still returns healthy status",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody: models.Response{
				Success: true,
				Message: "TailSwan control server is healthy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler()
			req := httptest.NewRequest(tt.method, "/health", http.NoBody)
			rec := httptest.NewRecorder()

			handler.Check(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			var resp models.Response
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success != tt.expectedBody.Success {
				t.Errorf("expected Success %v, got %v", tt.expectedBody.Success, resp.Success)
			}

			if resp.Message != tt.expectedBody.Message {
				t.Errorf("expected Message %q, got %q", tt.expectedBody.Message, resp.Message)
			}
		})
	}
}

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}
