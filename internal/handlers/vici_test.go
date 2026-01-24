package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klowdo/tailswan/internal/models"
)

const contentTypeJSON = "application/json"

func TestVICIHandler_ConnectionUp_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
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
			handler := &VICIHandler{}
			req := httptest.NewRequest(tt.method, "/api/vici/up", http.NoBody)
			rec := httptest.NewRecorder()

			handler.ConnectionUp(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestVICIHandler_ConnectionUp_InvalidJSON(t *testing.T) {
	handler := &VICIHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/vici/up", bytes.NewBufferString("invalid json"))
	rec := httptest.NewRecorder()

	handler.ConnectionUp(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != contentTypeJSON {
		t.Errorf("expected Content-Type %s, got %s", contentTypeJSON, contentType)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected Success to be false")
	}

	if resp.Message != "Invalid request" {
		t.Errorf("expected Message 'Invalid request', got %q", resp.Message)
	}
}

func TestVICIHandler_ConnectionUp_EmptyName(t *testing.T) {
	handler := &VICIHandler{}
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/vici/up", body)
	rec := httptest.NewRecorder()

	handler.ConnectionUp(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected Success to be false")
	}

	if resp.Message != "Connection name is required" {
		t.Errorf("expected Message 'Connection name is required', got %q", resp.Message)
	}

	if resp.Error != "name field cannot be empty" {
		t.Errorf("expected Error 'name field cannot be empty', got %q", resp.Error)
	}
}

func TestVICIHandler_ConnectionDown_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
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
			handler := &VICIHandler{}
			req := httptest.NewRequest(tt.method, "/api/vici/down", http.NoBody)
			rec := httptest.NewRecorder()

			handler.ConnectionDown(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestVICIHandler_ConnectionDown_InvalidJSON(t *testing.T) {
	handler := &VICIHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/vici/down", bytes.NewBufferString("{invalid}"))
	rec := httptest.NewRecorder()

	handler.ConnectionDown(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected Success to be false")
	}
}

func TestVICIHandler_ConnectionDown_EmptyName(t *testing.T) {
	handler := &VICIHandler{}
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/vici/down", body)
	rec := httptest.NewRecorder()

	handler.ConnectionDown(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected Success to be false")
	}

	if resp.Message != "Connection name is required" {
		t.Errorf("expected Message 'Connection name is required', got %q", resp.Message)
	}
}

func TestVICIHandler_ListConnections_MethodNotAllowed(t *testing.T) {
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
			handler := &VICIHandler{}
			req := httptest.NewRequest(tt.method, "/api/vici/connections", http.NoBody)
			rec := httptest.NewRecorder()

			handler.ListConnections(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestVICIHandler_ListSAs_MethodNotAllowed(t *testing.T) {
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
			handler := &VICIHandler{}
			req := httptest.NewRequest(tt.method, "/api/vici/sas", http.NoBody)
			rec := httptest.NewRecorder()

			handler.ListSAs(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestVICIHandler_Close_NilSession(t *testing.T) {
	handler := &VICIHandler{session: nil}
	err := handler.Close()
	if err != nil {
		t.Errorf("expected nil error for nil session, got %v", err)
	}
}

func TestVICIHandler_Session(t *testing.T) {
	handler := &VICIHandler{session: nil}
	if handler.Session() != nil {
		t.Error("expected nil session")
	}
}
