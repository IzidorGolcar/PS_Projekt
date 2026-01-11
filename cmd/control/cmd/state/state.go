package state

import (
	"github.com/spf13/cobra"
)

var (
	addr string
	Cmd  = &cobra.Command{
		Use:   "state",
		Short: "Get the current state of the RAFT node",
		Run:   run,
	}
)

func init() {
	Cmd.Flags().StringVarP(&addr, "addr", "a", "", "Address of the server")
	_ = Cmd.MarkFlagRequired("addr")
}
