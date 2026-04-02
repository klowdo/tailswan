package supervisor

import (
	"context"
	"fmt"

	"tailscale.com/client/local"

	"github.com/klowdo/tailswan/internal/swan"
)

func HealthCheck() error {
	client := &local.Client{}
	ctx := context.Background()

	status, err := client.StatusWithoutPeers(ctx)
	if err != nil {
		return fmt.Errorf("tailscaled not responding: %w", err)
	}

	if len(status.Health) > 0 {
		var problems string
		for warningCode, warningText := range status.Health {
			if problems != "" {
				problems += "; "
			}

			problems += fmt.Sprintf("%d: %s", warningCode, warningText)
		}
		return fmt.Errorf("tailscale health issues: %s", problems)
	}

	svc, err := swan.NewService()
	if err != nil {
		return fmt.Errorf("VICI not responding: %w", err)
	}
	defer svc.Close() //nolint:errcheck // best-effort cleanup

	if err := svc.Version(); err != nil {
		return fmt.Errorf("strongSwan not responding: %w", err)
	}

	return nil
}
