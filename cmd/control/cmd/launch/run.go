package launch

import (
	"fmt"
	"log"
	"seminarska/internal/control"

	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()

	fms := control.NewChainFSM()
	dataDir := fmt.Sprintf("data_%s", nodeId)

	r, err := control.SetupRaft(
		nodeId,
		raftAddr,
		dataDir,
		fms,
		bootstrap,
	)
	if err != nil {
		log.Fatal(err)
	}

	cfg := control.ChainConfig{
		LoggerPath:      logFile,
		DataExecutable:  dataExec,
		TargetNodeCount: 5,
	}
	manager := control.NewChainManager(ctx, cfg, fms, r, rpcAddr)

	go control.StartHTTP(httpAddr, r, fms)
	<-manager.Done()
}
