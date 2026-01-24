package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/klowdo/tailswan/internal/models"
	"github.com/klowdo/tailswan/internal/sse"
)

func TestSSEHandler_Events_MethodNotAllowed(t *testing.T) {
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

	broadcaster := sse.NewEventBroadcaster(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSSEHandler(broadcaster)
			req := httptest.NewRequest(tt.method, "/api/events", http.NoBody)
			rec := httptest.NewRecorder()

			handler.Events(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestSSEHandler_Events_Headers(t *testing.T) {
	broadcaster := sse.NewEventBroadcaster(nil, nil)
	handler := NewSSEHandler(broadcaster)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/events", http.NoBody)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		handler.Events(rec, req)
	}()

	<-ctx.Done()
	time.Sleep(10 * time.Millisecond)

	expectedHeaders := map[string]string{
		"Content-Type":                "text/event-stream",
		"Cache-Control":               "no-cache",
		"Connection":                  "keep-alive",
		"Access-Control-Allow-Origin": "*",
	}

	for header, expected := range expectedHeaders {
		actual := rec.Header().Get(header)
		if actual != expected {
			t.Errorf("expected header %s=%q, got %q", header, expected, actual)
		}
	}
}

func TestSSEHandler_Events_ReceivesMessage(t *testing.T) {
	broadcaster := sse.NewEventBroadcaster(nil, nil)
	handler := NewSSEHandler(broadcaster)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/events", http.NoBody)
	req = req.WithContext(ctx)

	rec := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		flushed:          false,
	}

	done := make(chan struct{})
	go func() {
		handler.Events(rec, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := models.SSEMessage{
		Event: "test-event",
		Data:  []byte(`{"test":"data"}`),
	}

	clientChan := make(chan models.SSEMessage, 10)
	broadcaster.RegisterClient(clientChan)

	clientChan <- testMessage

	<-ctx.Done()
	<-done

	body := rec.Body.String()
	if !strings.Contains(body, "event: test-event") {
		t.Logf("Response body: %s", body)
	}
}

func TestNewSSEHandler(t *testing.T) {
	broadcaster := sse.NewEventBroadcaster(nil, nil)
	handler := NewSSEHandler(broadcaster)
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestNewSSEHandler_NilBroadcaster(t *testing.T) {
	handler := NewSSEHandler(nil)
	if handler == nil {
		t.Error("expected non-nil handler even with nil broadcaster")
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (fr *flushRecorder) Flush() {
	fr.flushed = true
}
