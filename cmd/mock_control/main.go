package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"seminarska/internal/control/dataplane"
)

// **************************************
//          MOCK CONTROL SERVICE
// **************************************

func main() {
	log.SetOutput(os.Stdout)
	nodeExecPath := flag.String("path", "", "Path to the data node executable")
	flag.Parse()
	if *nodeExecPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := dataplane.ChainConfig{
		LoggerPath:      "/Users/izidor/Code/UNI/PS/seminarska/logs/dataplane.log",
		DataExecutable:  *nodeExecPath,
		TargetNodeCount: 5,
	}
	manager := dataplane.NewChainManager(ctx, cfg, ":8080")

	//service := NewMockControlService(*nodeExecPath, ":8080")

	fmt.Println("Listening on :8080")
	fmt.Println("Press Ctrl+C to exit")
	<-ctx.Done()
	fmt.Println("Stopping...")
	<-manager.Done()
}
