package swan

import (
	"context"
	"fmt"

	"github.com/strongswan/govici/vici"
)

type Service struct {
	session *vici.Session
}

func NewService() (*Service, error) {
	session, err := vici.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create VICI session: %w", err)
	}
	return &Service{session: session}, nil
}

func (s *Service) Close() error {
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}

func (s *Service) Session() *vici.Session {
	return s.session
}

func (s *Service) LoadAll() error {
	msg := vici.NewMessage()
	for _, cmd := range []string{"load-conn", "load-cert", "load-authority", "load-pool", "load-secret"} {
		if _, err := s.session.Call(context.Background(), cmd, msg); err != nil {
			return fmt.Errorf("VICI %s: %w", cmd, err)
		}
	}
	return nil
}

func (s *Service) Initiate(child string) error {
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
	msg := vici.NewMessage()
	if _, err := s.session.Call(context.Background(), "version", msg); err != nil {
		return fmt.Errorf("VICI version: %w", err)
	}
	return nil
}
