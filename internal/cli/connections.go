package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klowdo/tailswan/internal/swan"
)

func NewConnectionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "connections",
		Aliases: []string{"conns"},
		Short:   "List strongSwan connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := swan.NewService()
			if err != nil {
				return fmt.Errorf("VICI connection failed: %w", err)
			}
			defer svc.Close() //nolint:errcheck // best-effort cleanup

			connections, err := svc.ListConnections()
			if err != nil {
				return fmt.Errorf("failed to list connections: %w", err)
			}

			for _, conn := range connections {
				for name, details := range conn {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %v\n", name, details); err != nil {
						return fmt.Errorf("failed to write output: %w", err)
					}
				}
			}
			return nil
		},
	}
}
