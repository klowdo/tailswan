package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "=== Tailscale Status ==="); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			tailscaleCmd := exec.Command("tailscale", "status")
			tailscaleCmd.Stdout = cmd.OutOrStdout()
			tailscaleCmd.Stderr = cmd.ErrOrStderr()
			if err := tailscaleCmd.Run(); err != nil {
				return fmt.Errorf("tailscale status failed: %w", err)
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\n=== strongSwan Connections ==="); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			swanctlCmd := exec.Command("swanctl", "--list-conns")
			swanctlCmd.Stdout = cmd.OutOrStdout()
			swanctlCmd.Stderr = cmd.ErrOrStderr()
			if err := swanctlCmd.Run(); err != nil {
				return fmt.Errorf("swanctl failed: %w", err)
			}

			return nil
		},
	}
}
