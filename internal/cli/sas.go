package cli

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewSAsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sas",
		Short: "List security associations",
		RunE: func(cmd *cobra.Command, args []string) error {
			swanctlCmd := exec.Command("swanctl", "--list-sas")
			swanctlCmd.Stdout = cmd.OutOrStdout()
			swanctlCmd.Stderr = cmd.ErrOrStderr()
			return swanctlCmd.Run()
		},
	}
}
