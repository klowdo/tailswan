package supervisor

import (
	"context"
	"fmt"
	"os/exec"

	"tailscale.com/client/local"
)

func HealthCheck() error {
	client := &local.Client{}
	ctx := context.Background()

	status, err := client.StatusWithoutPeers(ctx)
	if err != nil {
		return fmt.Errorf("tailscaled not responding: %w", err)
	}

	// Health contains health check problems. Empty means everything is good.
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

	if err := exec.Command("swanctl", "--stats").Run(); err != nil {
		return fmt.Errorf("charon not responding via vici: %w", err)
	}

	return nil
}
