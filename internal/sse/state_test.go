package sse

import (
	"sync"
	"testing"
)

func TestNewStateTracker(t *testing.T) {
	st := NewStateTracker()
	if st == nil {
		t.Error("NewStateTracker() returned nil")
	}
}

func TestStateTrackerHasChangedFirstCall(t *testing.T) {
	st := NewStateTracker()

	tests := []struct {
		data interface{}
		key  string
	}{
		{key: "sas", data: map[string]string{"key": "value"}},
		{key: "peers", data: []string{"peer1", "peer2"}},
		{key: "connections", data: map[string]int{"count": 5}},
		{key: "node", data: struct{ Name string }{"node1"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if !st.HasChanged(tt.key, tt.data) {
				t.Errorf("HasChanged(%q) should return true on first call", tt.key)
			}
		})
	}
}

func TestStateTrackerHasChangedSameData(t *testing.T) {
	st := NewStateTracker()
	data := map[string]string{"key": "value"}

	st.HasChanged("sas", data)

	if st.HasChanged("sas", data) {
		t.Error("HasChanged() should return false for unchanged data")
	}
}

func TestStateTrackerHasChangedDifferentData(t *testing.T) {
	st := NewStateTracker()

	st.HasChanged("sas", map[string]string{"key": "value1"})

	if !st.HasChanged("sas", map[string]string{"key": "value2"}) {
		t.Error("HasChanged() should return true for changed data")
	}
}

func TestStateTrackerUnknownKey(t *testing.T) {
	st := NewStateTracker()

	if st.HasChanged("unknown", "data") {
		t.Error("HasChanged() should return false for unknown key")
	}
}

func TestStateTrackerConcurrency(t *testing.T) {
	st := NewStateTracker()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st.HasChanged("sas", map[string]int{"value": i})
			st.HasChanged("peers", map[string]int{"value": i})
		}(i)
	}

	wg.Wait()
}

func TestComputeHash(t *testing.T) {
	hash1 := computeHash(map[string]string{"key": "value"})
	hash2 := computeHash(map[string]string{"key": "value"})
	hash3 := computeHash(map[string]string{"key": "different"})

	if hash1 == "" {
		t.Error("computeHash() returned empty string")
	}

	if hash1 != hash2 {
		t.Error("computeHash() should return same hash for same data")
	}

	if hash1 == hash3 {
		t.Error("computeHash() should return different hash for different data")
	}
}

func TestComputeHashDifferentTypes(t *testing.T) {
	tests := []struct {
		data interface{}
		name string
	}{
		{name: "string", data: "test"},
		{name: "int", data: 42},
		{name: "slice", data: []int{1, 2, 3}},
		{name: "map", data: map[string]int{"a": 1}},
		{name: "struct", data: struct{ X int }{X: 1}},
		{name: "nil", data: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := computeHash(tt.data)
			if hash == "" && tt.data != nil {
				t.Errorf("computeHash(%v) returned empty string", tt.data)
			}
		})
	}
}
