package swan

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"github.com/strongswan/govici/vici"
)

type Service struct {
	session *vici.Session
	mu      sync.Mutex
}

func NewService() (*Service, error) {
	session, err := vici.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create VICI session: %w", err)
	}
	return &Service{session: session}, nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}

func (s *Service) Session() *vici.Session {
	return s.session
}

func (s *Service) LoadAll() error {
	slog.Info("Loading swanctl configuration")
	cmd := exec.Command("swanctl", "--load-all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("swanctl --load-all: %w", err)
	}
	return nil
}

func (s *Service) Initiate(child string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := vici.NewMessage()
	if err := msg.Set("child", child); err != nil {
		return fmt.Errorf("set child field: %w", err)
	}
	if _, err := s.session.Call(context.Background(), "initiate", msg); err != nil {
		return fmt.Errorf("initiate %s: %w", child, err)
	}
	return nil
}

func (s *Service) Terminate(ike string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := vici.NewMessage()
	if err := msg.Set("ike", ike); err != nil {
		return fmt.Errorf("set ike field: %w", err)
	}
	if _, err := s.session.Call(context.Background(), "terminate", msg); err != nil {
		return fmt.Errorf("terminate %s: %w", ike, err)
	}
	return nil
}

func (s *Service) ListConnections() ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := vici.NewMessage()
	connections := make([]map[string]any, 0)
	for m, err := range s.session.CallStreaming(context.Background(), "list-conns", "list-conn", msg) {
		if err != nil {
			return nil, fmt.Errorf("list connections: %w", err)
		}
		connMap := make(map[string]any)
		for _, key := range m.Keys() {
			connMap[key] = m.Get(key)
		}
		connections = append(connections, connMap)
	}
	return connections, nil
}

func (s *Service) ListSAs() ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := vici.NewMessage()
	sas := make([]map[string]any, 0)
	for m, err := range s.session.CallStreaming(context.Background(), "list-sas", "list-sa", msg) {
		if err != nil {
			return nil, fmt.Errorf("list SAs: %w", err)
		}
		saMap := make(map[string]any)
		for _, key := range m.Keys() {
			saMap[key] = m.Get(key)
		}
		sas = append(sas, saMap)
	}
	return sas, nil
}

func (s *Service) Version() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := vici.NewMessage()
	if _, err := s.session.Call(context.Background(), "version", msg); err != nil {
		return fmt.Errorf("VICI version: %w", err)
	}
	return nil
}
