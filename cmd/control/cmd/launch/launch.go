package launch

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	nodeId          string
	raftAddr        string
	httpAddr        string
	rpcAddr         string
	dataExec        string
	logFile         string
	bootstrap       bool
	targetDataNodes int
	Cmd             = &cobra.Command{
		Use:   "launch",
		Short: "Launch a new data service node",
		Run:   run,
	}
)

func init() {
	Cmd.Flags().StringVar(&nodeId, "node-id", "", "Node ID")
	Cmd.Flags().StringVar(&raftAddr, "raft-addr", "", "Raft address")
	Cmd.Flags().StringVar(&httpAddr, "http-addr", "", "HTTP address")
	Cmd.Flags().StringVar(&rpcAddr, "rpc-addr", "", "RPC address")
	Cmd.Flags().StringVar(&dataExec, "data-exec", "", "Data execution path")
	Cmd.Flags().BoolVar(&bootstrap, "bootstrap", false, "Bootstrap flag")
	Cmd.Flags().StringVar(&logFile, "logs", os.DevNull, "Log file")
	Cmd.Flags().IntVar(&targetDataNodes, "target-nodes", 5, "Target number of data nodes")

	_ = Cmd.MarkFlagRequired("node-id")
	_ = Cmd.MarkFlagRequired("raft-addr")
	_ = Cmd.MarkFlagRequired("http-addr")
	_ = Cmd.MarkFlagRequired("rpc-addr")
	_ = Cmd.MarkFlagRequired("data-exec")
}
