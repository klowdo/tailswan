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
			fmt.Fprintln(cmd.OutOrStdout(), "=== Tailscale Status ===")
			tailscaleCmd := exec.Command("tailscale", "status")
			tailscaleCmd.Stdout = cmd.OutOrStdout()
			tailscaleCmd.Stderr = cmd.ErrOrStderr()
			tailscaleCmd.Run()

			fmt.Fprintln(cmd.OutOrStdout(), "\n=== strongSwan Connections ===")
			swanctlCmd := exec.Command("swanctl", "--list-conns")
			swanctlCmd.Stdout = cmd.OutOrStdout()
			swanctlCmd.Stderr = cmd.ErrOrStderr()
			swanctlCmd.Run()

			return nil
		},
	}
}
