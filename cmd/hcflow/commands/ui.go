// Package commands — ui subcommand.
package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/web"
)

func newUICmd() *cobra.Command {
	var noBrowser bool

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Open the hcflow web dashboard",
		Long: `Start a local HTTP server and open the hcflow dashboard in your browser.

The server binds to 127.0.0.1 on a random available port.
Press Ctrl+C to stop.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			srv := web.New(dir)
			addr, err := srv.Start(!noBrowser)
			if err != nil {
				return fmt.Errorf("starting server: %w", err)
			}

			fmt.Printf("%s hcflow UI running at %s\n", checkMark(), addr)
			if !noBrowser {
				fmt.Println(dim("  Opening browser…"))
			}
			fmt.Println(dim("  Press Ctrl+C to stop"))

			// Wait for signal
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			<-quit

			fmt.Println("\nShutting down…")
			return srv.Shutdown(context.Background())
		},
	}

	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "do not open browser automatically")
	return cmd
}
