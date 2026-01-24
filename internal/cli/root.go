package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func NewRootCmd(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tailswan",
		Short: "TailSwan - IPsec & Tailscale VPN Supervisor",
		Long:  `TailSwan is a unified supervisor for managing strongSwan IPsec and Tailscale VPN connections.`,
	}

	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	rootCmd.AddCommand(
		NewHealthCheckCmd(),
		NewStatusCmd(),
		NewConnectionsCmd(),
		NewSAsCmd(),
		NewStartCmd(),
		NewStopCmd(),
		NewReloadCmd(),
	)

	return rootCmd
}
