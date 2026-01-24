package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func jsonEqual(a, b []byte) bool {
	var aVal, bVal interface{}
	if err := json.Unmarshal(a, &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(aVal, bVal)
}

func TestConnectionRequest_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    ConnectionRequest
		expected string
	}{
		{
			name:     "simple name",
			input:    ConnectionRequest{Name: "test-connection"},
			expected: `{"name":"test-connection"}`,
		},
		{
			name:     "empty name",
			input:    ConnectionRequest{Name: ""},
			expected: `{"name":""}`,
		},
		{
			name:     "name with special characters",
			input:    ConnectionRequest{Name: "test-conn_123"},
			expected: `{"name":"test-conn_123"}`,
		},
		{
			name:     "name with unicode",
			input:    ConnectionRequest{Name: "connection-\u00e9"},
			expected: `{"name":"connection-é"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(result, []byte(tt.expected)) {
				t.Errorf("got %s, want %s", string(result), tt.expected)
			}
		})
	}
}

func TestConnectionRequest_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ConnectionRequest
		wantError bool
	}{
		{
			name:     "valid json",
			input:    `{"name":"my-connection"}`,
			expected: ConnectionRequest{Name: "my-connection"},
		},
		{
			name:     "empty name value",
			input:    `{"name":""}`,
			expected: ConnectionRequest{Name: ""},
		},
		{
			name:     "extra fields ignored",
			input:    `{"name":"conn","extra":"field"}`,
			expected: ConnectionRequest{Name: "conn"},
		},
		{
			name:     "missing name field",
			input:    `{}`,
			expected: ConnectionRequest{Name: ""},
		},
		{
			name:      "invalid json",
			input:     `{invalid}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ConnectionRequest
			err := json.Unmarshal([]byte(tt.input), &result)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestResponse_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		input    Response
	}{
		{
			name:     "success without error",
			input:    Response{Success: true, Message: "operation completed"},
			expected: `{"success":true,"message":"operation completed"}`,
		},
		{
			name:     "failure with error",
			input:    Response{Success: false, Message: "operation failed", Error: "connection timeout"},
			expected: `{"success":false,"message":"operation failed","error":"connection timeout"}`,
		},
		{
			name:     "error omitted when empty",
			input:    Response{Success: true, Message: "ok", Error: ""},
			expected: `{"success":true,"message":"ok"}`,
		},
		{
			name:     "all fields empty except success",
			input:    Response{Success: false, Message: "", Error: ""},
			expected: `{"success":false,"message":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(result, []byte(tt.expected)) {
				t.Errorf("got %s, want %s", string(result), tt.expected)
			}
		})
	}
}

func TestResponse_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Response
		wantError bool
	}{
		{
			name:     "success response",
			input:    `{"success":true,"message":"done"}`,
			expected: Response{Success: true, Message: "done"},
		},
		{
			name:     "error response",
			input:    `{"success":false,"message":"failed","error":"timeout"}`,
			expected: Response{Success: false, Message: "failed", Error: "timeout"},
		},
		{
			name:     "missing optional error field",
			input:    `{"success":true,"message":"ok"}`,
			expected: Response{Success: true, Message: "ok", Error: ""},
		},
		{
			name:      "invalid json",
			input:     `not json`,
			wantError: true,
		},
		{
			name:      "wrong type for success",
			input:     `{"success":"yes","message":"ok"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Response
			err := json.Unmarshal([]byte(tt.input), &result)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestConnectionsResponse_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		input    ConnectionsResponse
	}{
		{
			name: "with connections",
			input: ConnectionsResponse{
				Success: true,
				Connections: []map[string]interface{}{
					{"id": "conn1", "status": "active"},
					{"id": "conn2", "status": "inactive"},
				},
			},
			expected: `{"success":true,"connections":[{"id":"conn1","status":"active"},{"id":"conn2","status":"inactive"}]}`,
		},
		{
			name: "empty connections",
			input: ConnectionsResponse{
				Success:     true,
				Connections: []map[string]interface{}{},
			},
			expected: `{"success":true,"connections":[]}`,
		},
		{
			name: "nil connections",
			input: ConnectionsResponse{
				Success:     false,
				Connections: nil,
			},
			expected: `{"success":false,"connections":null}`,
		},
		{
			name: "connection with nested data",
			input: ConnectionsResponse{
				Success: true,
				Connections: []map[string]interface{}{
					{"id": "conn1", "config": map[string]interface{}{"port": float64(443)}},
				},
			},
			expected: `{"success":true,"connections":[{"config":{"port":443},"id":"conn1"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(result, []byte(tt.expected)) {
				t.Errorf("got %s, want %s", string(result), tt.expected)
			}
		})
	}
}

func TestConnectionsResponse_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ConnectionsResponse
		wantError bool
	}{
		{
			name:  "valid response",
			input: `{"success":true,"connections":[{"id":"conn1"}]}`,
			expected: ConnectionsResponse{
				Success:     true,
				Connections: []map[string]interface{}{{"id": "conn1"}},
			},
		},
		{
			name:  "empty connections array",
			input: `{"success":true,"connections":[]}`,
			expected: ConnectionsResponse{
				Success:     true,
				Connections: []map[string]interface{}{},
			},
		},
		{
			name:  "null connections",
			input: `{"success":false,"connections":null}`,
			expected: ConnectionsResponse{
				Success:     false,
				Connections: nil,
			},
		},
		{
			name:      "invalid json",
			input:     `{broken`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ConnectionsResponse
			err := json.Unmarshal([]byte(tt.input), &result)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.expected.Success {
				t.Errorf("Success: got %v, want %v", result.Success, tt.expected.Success)
			}
			if !reflect.DeepEqual(result.Connections, tt.expected.Connections) {
				t.Errorf("Connections: got %+v, want %+v", result.Connections, tt.expected.Connections)
			}
		})
	}
}

func TestSAsResponse_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		input    SAsResponse
	}{
		{
			name: "with security associations",
			input: SAsResponse{
				Success: true,
				SAs: []map[string]interface{}{
					{"name": "sa1", "state": "established"},
					{"name": "sa2", "state": "connecting"},
				},
			},
			expected: `{"success":true,"sas":[{"name":"sa1","state":"established"},{"name":"sa2","state":"connecting"}]}`,
		},
		{
			name: "empty sas",
			input: SAsResponse{
				Success: true,
				SAs:     []map[string]interface{}{},
			},
			expected: `{"success":true,"sas":[]}`,
		},
		{
			name: "nil sas",
			input: SAsResponse{
				Success: false,
				SAs:     nil,
			},
			expected: `{"success":false,"sas":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(result, []byte(tt.expected)) {
				t.Errorf("got %s, want %s", string(result), tt.expected)
			}
		})
	}
}

func TestSAsResponse_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  SAsResponse
		wantError bool
	}{
		{
			name:  "valid response",
			input: `{"success":true,"sas":[{"name":"sa1","state":"established"}]}`,
			expected: SAsResponse{
				Success: true,
				SAs:     []map[string]interface{}{{"name": "sa1", "state": "established"}},
			},
		},
		{
			name:  "empty sas array",
			input: `{"success":true,"sas":[]}`,
			expected: SAsResponse{
				Success: true,
				SAs:     []map[string]interface{}{},
			},
		},
		{
			name:  "null sas",
			input: `{"success":false,"sas":null}`,
			expected: SAsResponse{
				Success: false,
				SAs:     nil,
			},
		},
		{
			name:      "invalid json",
			input:     `{invalid}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SAsResponse
			err := json.Unmarshal([]byte(tt.input), &result)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.expected.Success {
				t.Errorf("Success: got %v, want %v", result.Success, tt.expected.Success)
			}
			if !reflect.DeepEqual(result.SAs, tt.expected.SAs) {
				t.Errorf("SAs: got %+v, want %+v", result.SAs, tt.expected.SAs)
			}
		})
	}
}

func TestSSEMessage_Fields(t *testing.T) {
	tests := []struct {
		name          string
		msg           SSEMessage
		expectedEvent string
		expectedData  []byte
	}{
		{
			name:          "standard message",
			msg:           SSEMessage{Event: "connection", Data: []byte(`{"status":"connected"}`)},
			expectedEvent: "connection",
			expectedData:  []byte(`{"status":"connected"}`),
		},
		{
			name:          "empty event",
			msg:           SSEMessage{Event: "", Data: []byte("data")},
			expectedEvent: "",
			expectedData:  []byte("data"),
		},
		{
			name:          "nil data",
			msg:           SSEMessage{Event: "event", Data: nil},
			expectedEvent: "event",
			expectedData:  nil,
		},
		{
			name:          "empty data",
			msg:           SSEMessage{Event: "event", Data: []byte{}},
			expectedEvent: "event",
			expectedData:  []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Event != tt.expectedEvent {
				t.Errorf("Event: got %s, want %s", tt.msg.Event, tt.expectedEvent)
			}
			if !reflect.DeepEqual(tt.msg.Data, tt.expectedData) {
				t.Errorf("Data: got %v, want %v", tt.msg.Data, tt.expectedData)
			}
		})
	}
}

func TestResponse_OmitEmpty(t *testing.T) {
	response := Response{Success: true, Message: "ok", Error: ""}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := unmarshaled["error"]; exists {
		t.Error("error field should be omitted when empty")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	t.Run("ConnectionRequest", func(t *testing.T) {
		original := ConnectionRequest{Name: "test-connection"}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var result ConnectionRequest
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if result != original {
			t.Errorf("round trip failed: got %+v, want %+v", result, original)
		}
	})

	t.Run("Response", func(t *testing.T) {
		original := Response{Success: false, Message: "error occurred", Error: "timeout"}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var result Response
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if result != original {
			t.Errorf("round trip failed: got %+v, want %+v", result, original)
		}
	})

	t.Run("ConnectionsResponse", func(t *testing.T) {
		original := ConnectionsResponse{
			Success:     true,
			Connections: []map[string]interface{}{{"id": "test"}},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var result ConnectionsResponse
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if result.Success != original.Success {
			t.Errorf("Success: got %v, want %v", result.Success, original.Success)
		}
		if !reflect.DeepEqual(result.Connections, original.Connections) {
			t.Errorf("Connections: got %+v, want %+v", result.Connections, original.Connections)
		}
	})

	t.Run("SAsResponse", func(t *testing.T) {
		original := SAsResponse{
			Success: true,
			SAs:     []map[string]interface{}{{"name": "sa1"}},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var result SAsResponse
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if result.Success != original.Success {
			t.Errorf("Success: got %v, want %v", result.Success, original.Success)
		}
		if !reflect.DeepEqual(result.SAs, original.SAs) {
			t.Errorf("SAs: got %+v, want %+v", result.SAs, original.SAs)
		}
	})
}
