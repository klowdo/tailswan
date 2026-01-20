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

	_, err := client.Status(ctx)
	if err != nil {
		return fmt.Errorf("tailscaled not responding: %w", err)
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
