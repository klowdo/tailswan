package cli

import (
	"fmt"

	"github.com/klowdo/tailswan/internal/supervisor"
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <connection>",
		Short: "Terminate a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := args[0]
			sw := &supervisor.SwanService{}
			if err := sw.Terminate(conn); err != nil {
				return fmt.Errorf("failed to stop connection %s: %w", conn, err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Connection '%s' terminated\n", conn); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			return nil
		},
	}
}
