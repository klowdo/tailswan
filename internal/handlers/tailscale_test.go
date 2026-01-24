package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTailscaleHandler_Status_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH method not allowed",
			method:         http.MethodPatch,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTailscaleHandler()
			req := httptest.NewRequest(tt.method, "/api/tailscale/status", http.NoBody)
			rec := httptest.NewRecorder()

			handler.Status(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestTailscaleHandler_Peers_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTailscaleHandler()
			req := httptest.NewRequest(tt.method, "/api/tailscale/peers", http.NoBody)
			rec := httptest.NewRecorder()

			handler.Peers(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestTailscaleHandler_ServeStatus_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTailscaleHandler()
			req := httptest.NewRequest(tt.method, "/api/tailscale/serve", http.NoBody)
			rec := httptest.NewRecorder()

			handler.ServeStatus(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestTailscaleHandler_WhoIs_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTailscaleHandler()
			req := httptest.NewRequest(tt.method, "/api/tailscale/whois", http.NoBody)
			rec := httptest.NewRecorder()

			handler.WhoIs(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestNewTailscaleHandler(t *testing.T) {
	handler := NewTailscaleHandler()
	if handler == nil {
		t.Error("expected non-nil handler")
	}
	if handler.LocalClient() == nil {
		t.Error("expected non-nil LocalClient")
	}
}

func TestNewTailscaleHandlerWithClient(t *testing.T) {
	handler := NewTailscaleHandlerWithClient(nil)
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestTailscaleHandler_SetClient(t *testing.T) {
	handler := NewTailscaleHandler()
	originalClient := handler.LocalClient()

	handler.SetClient(nil)
	if handler.LocalClient() != nil {
		t.Error("expected nil client after SetClient(nil)")
	}

	handler.SetClient(originalClient)
	if handler.LocalClient() != originalClient {
		t.Error("expected client to be restored")
	}
}

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		data           interface{}
		name           string
		status         int
		expectedStatus int
	}{
		{
			name:           "respond with 200 OK",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "respond with 400 Bad Request",
			status:         http.StatusBadRequest,
			data:           map[string]string{"error": "bad request"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "respond with 500 Internal Server Error",
			status:         http.StatusInternalServerError,
			data:           map[string]string{"error": "internal error"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondJSON(rec, tt.status, tt.data)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			var result map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
		})
	}
}
