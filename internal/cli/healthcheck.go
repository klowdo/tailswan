package cli

import (
	"fmt"

	"github.com/klowdo/tailswan/internal/supervisor"
	"github.com/spf13/cobra"
)

func NewHealthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "healthcheck",
		Short: "Check if all services are healthy",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := supervisor.HealthCheck(); err != nil {
				return fmt.Errorf("health check failed: %w", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "All services healthy"); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			return nil
		},
	}
}
