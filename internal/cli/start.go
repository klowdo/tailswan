package cli

import (
	"fmt"

	"github.com/klowdo/tailswan/internal/supervisor"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <connection>",
		Short: "Initiate a connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := args[0]
			sw := &supervisor.SwanService{}
			if err := sw.Initiate(conn); err != nil {
				return fmt.Errorf("failed to start connection %s: %w", conn, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Connection '%s' initiated\n", conn)
			return nil
		},
	}
}
