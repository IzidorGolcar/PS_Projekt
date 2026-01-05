package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

// **************************************
//          MOCK CONTROL SERVICE
// **************************************

func main() {
	nodeExecPath := flag.String("path", "", "Path to the data node executable")
	flag.Parse()
	if *nodeExecPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	service := NewMockControlService(*nodeExecPath, ":8080")
	fmt.Println("Listening on :8080")
	fmt.Println("Press Ctrl+C to exit")
	<-ctx.Done()
	fmt.Println("Stopping...")
	service.Shutdown()
}
