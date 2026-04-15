package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klowdo/tailswan/internal/swan"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <connection>",
		Short: "Terminate a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := swan.NewService()
			if err != nil {
				return fmt.Errorf("VICI connection failed: %w", err)
			}
			defer svc.Close() //nolint:errcheck // best-effort cleanup

			conn := args[0]
			if err := svc.Terminate(conn); err != nil {
				return fmt.Errorf("failed to stop connection %s: %w", conn, err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Connection '%s' terminated\n", conn); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			return nil
		},
	}
}
