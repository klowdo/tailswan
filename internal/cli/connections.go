package cli

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewConnectionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "connections",
		Aliases: []string{"conns"},
		Short:   "List strongSwan connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			swanctlCmd := exec.Command("swanctl", "--list-conns")
			swanctlCmd.Stdout = cmd.OutOrStdout()
			swanctlCmd.Stderr = cmd.ErrOrStderr()
			return swanctlCmd.Run()
		},
	}
}
