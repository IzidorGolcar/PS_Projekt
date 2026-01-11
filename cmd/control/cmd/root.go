package cmd

import (
	"context"
	"os"
	"os/signal"
	"seminarska/cmd/control/cmd/launch"
	"seminarska/cmd/control/cmd/link"
	"seminarska/cmd/control/cmd/state"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "control",
	Short: "Distributed control plane",
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(launch.Cmd)
	rootCmd.AddCommand(link.Cmd)
	rootCmd.AddCommand(state.Cmd)
}
