package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"tailscale.com/client/local"

	"github.com/klowdo/tailswan/internal/swan"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if err := printTailscaleStatus(out); err != nil {
				return err
			}

			return printSwanStatus(out)
		},
	}
}

func printTailscaleStatus(out io.Writer) error {
	if _, err := fmt.Fprintln(out, "=== Tailscale Status ==="); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	tsClient := &local.Client{}
	status, err := tsClient.Status(context.Background())
	if err != nil {
		return fmt.Errorf("tailscale status failed: %w", err)
	}

	if _, err := fmt.Fprintf(out, "Backend: %s\n", status.BackendState); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if status.Self != nil {
		if _, err := fmt.Fprintf(out, "Hostname: %s\nDNS: %s\nIPs: %v\n",
			status.Self.HostName, status.Self.DNSName, status.Self.TailscaleIPs); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	for _, peer := range status.Peer {
		if _, err := fmt.Fprintf(out, "  %s (%s) online=%v\n",
			peer.HostName, peer.DNSName, peer.Online); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

func printSwanStatus(out io.Writer) error {
	if _, err := fmt.Fprintln(out, "\n=== strongSwan Connections ==="); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

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
			if _, err := fmt.Fprintf(out, "%s: %v\n", name, details); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
	}

	return nil
}
