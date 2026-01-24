package sse

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
)

type StateTracker struct {
	lastSAsHash         string
	lastPeersHash       string
	lastConnectionsHash string
	lastNodeHash        string
	mu                  sync.RWMutex
}

func NewStateTracker() *StateTracker {
	return &StateTracker{}
}

func (st *StateTracker) HasChanged(key string, newData interface{}) bool {
	newHash := computeHash(newData)

	st.mu.Lock()
	defer st.mu.Unlock()

	var lastHash *string
	switch key {
	case "sas":
		lastHash = &st.lastSAsHash
	case "peers":
		lastHash = &st.lastPeersHash
	case "connections":
		lastHash = &st.lastConnectionsHash
	case "node":
		lastHash = &st.lastNodeHash
	default:
		return false
	}

	if *lastHash != newHash {
		*lastHash = newHash
		return true
	}
	return false
}

func computeHash(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}
