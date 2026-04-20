package viciconn

import (
	"context"
	"log/slog"

	"github.com/strongswan/govici/vici"
)

func Build(session *vici.Session, configured []string) []map[string]interface{} {
	byName := make(map[string]map[string]interface{})
	order := make([]string, 0, len(configured))

	for _, name := range configured {
		byName[name] = map[string]interface{}{"loaded": false}
		order = append(order, name)
	}

	if session != nil {
		msg := vici.NewMessage()
		for m, err := range session.CallStreaming(context.Background(), "list-conns", "list-conn", msg) {
			if err != nil {
				slog.Info("Error fetching connections", "error", err)
				break
			}
			for _, name := range m.Keys() {
				details := map[string]interface{}{"loaded": true}
				if sub, ok := m.Get(name).(*vici.Message); ok {
					for _, k := range sub.Keys() {
						details[k] = sub.Get(k)
					}
				}
				if _, exists := byName[name]; !exists {
					order = append(order, name)
				}
				byName[name] = details
			}
		}
	}

	result := make([]map[string]interface{}, 0, len(order))
	for _, name := range order {
		result = append(result, map[string]interface{}{name: byName[name]})
	}
	return result
}
