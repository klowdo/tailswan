package supervisor

import (
	"context"
	"fmt"
	"os/exec"

	"tailscale.com/client/tailscale"
)

func HealthCheck() error {
	client := &tailscale.LocalClient{}
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

	cmd := exec.Command("swanctl", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("swanctl not responding: %w", err)
	}

	cmd = exec.Command("pgrep", "-x", "charon")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("charon process not found: %w", err)
	}

	return nil
}
