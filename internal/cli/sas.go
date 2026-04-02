package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/klowdo/tailswan/internal/swan"
)

func NewSAsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sas",
		Short: "List security associations",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := swan.NewService()
			if err != nil {
				return fmt.Errorf("VICI connection failed: %w", err)
			}
			defer svc.Close() //nolint:errcheck // best-effort cleanup

			sas, err := svc.ListSAs()
			if err != nil {
				return fmt.Errorf("failed to list SAs: %w", err)
			}

			for _, sa := range sas {
				for name, details := range sa {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %v\n", name, details); err != nil {
						return fmt.Errorf("failed to write output: %w", err)
					}
				}
			}
			return nil
		},
	}
}
