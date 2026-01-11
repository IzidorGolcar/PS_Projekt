package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"seminarska/internal/control"
)

func main() {
	var (
		nodeID       = flag.String("id", "", "node id")
		raftAddr     = flag.String("raft", "", "raft address")
		httpAddr     = flag.String("http", "", "http address")
		rpcAddr      = flag.String("rpc", "", "rpc address")
		dataDir      = flag.String("data", "", "data dir")
		nodeExecPath = flag.String("node-exec", "", "Path to the data node executable")
		bootstrap    = flag.Bool("bootstrap", false, "bootstrap cluster")
	)
	flag.Parse()

	if *nodeID == "" || *raftAddr == "" || *httpAddr == "" || *dataDir == "" {
		log.Fatal("missing required flags")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fms := control.NewChainFSM()

	r, err := control.SetupRaft(
		*nodeID,
		*raftAddr,
		*dataDir,
		fms,
		*bootstrap,
	)
	if err != nil {
		log.Fatal(err)
	}

	cfg := control.ChainConfig{
		LoggerPath:      "/Users/izidor/Code/UNI/PS/seminarska/logs/dataplane.log",
		DataExecutable:  *nodeExecPath,
		TargetNodeCount: 5,
	}
	manager := control.NewChainManager(ctx, cfg, fms, r, *rpcAddr)

	go control.StartHTTP(*httpAddr, r)
	<-manager.Done()
}
