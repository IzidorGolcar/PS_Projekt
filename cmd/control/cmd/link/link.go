package link

import (
	"github.com/spf13/cobra"
)

var (
	srcNodeAddr    string
	targetNodeAddr string
	targetNodeId   string

	Cmd = &cobra.Command{
		Use:   "link",
		Short: "Add a new node to the RAFT cluster",
		Run:   run,
	}
)

func init() {
	Cmd.Flags().StringVarP(&srcNodeAddr, "src", "s", "", "source node address")
	Cmd.Flags().StringVarP(&targetNodeAddr, "target", "t", "", "target node address")
	Cmd.Flags().StringVarP(&targetNodeId, "target-id", "i", "", "target node ID")

	_ = Cmd.MarkFlagRequired("src")
	_ = Cmd.MarkFlagRequired("target")
	_ = Cmd.MarkFlagRequired("target-id")
}
