package supervisor

import (
	"fmt"
	"os/exec"
)

func HealthCheck() error {
	cmd := exec.Command("tailscale", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tailscaled not responding: %w", err)
	}

	cmd = exec.Command("swanctl", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("swanctl not responding: %w", err)
	}

	cmd = exec.Command("pgrep", "-x", "charon")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("charon process not found: %w", err)
	}

	return nil
}
