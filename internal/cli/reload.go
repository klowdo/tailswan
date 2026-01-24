package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klowdo/tailswan/internal/supervisor"
)

func NewReloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reload",
		Short: "Reload strongSwan configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			sw := &supervisor.SwanService{}
			if err := sw.Reload(); err != nil {
				return fmt.Errorf("failed to reload configuration: %w", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Configuration reloaded"); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			return nil
		},
	}
}
