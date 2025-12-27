package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"seminarska/internal/raft"
)

func main() {
	// Parse command line flags
	nodeID := flag.String("id", "", "Node ID (required)")
	addr := flag.String("addr", ":7000", "Address to listen on for Raft RPCs")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (e.g., :7001,:7002)")
	peerIDs := flag.String("peer-ids", "", "Comma-separated list of peer IDs (e.g., node2,node3)")
	flag.Parse()

	if *nodeID == "" {
		fmt.Println("Error: -id is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse peers
	var peerAddrs []string
	var peerIDList []string
	if *peers != "" {
		peerAddrs = strings.Split(*peers, ",")
	}
	if *peerIDs != "" {
		peerIDList = strings.Split(*peerIDs, ",")
	}

	if len(peerAddrs) != len(peerIDList) {
		fmt.Println("Error: number of peers and peer-ids must match")
		os.Exit(1)
	}

	// Configure logger
	log.Default().SetOutput(os.Stdout)
	log.Default().SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.Default().SetPrefix(fmt.Sprintf("[ControlPlane %s] ", *nodeID))

	// Create context with cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Starting control plane node")
	log.Printf("Node ID: %s", *nodeID)
	log.Printf("Listen address: %s", *addr)
	log.Printf("Peers: %v", peerAddrs)

	// Create cluster manager (will be used as apply function)
	var clusterManager *raft.ClusterManager

	// Create Raft node
	raftNode := raft.NewNode(ctx, *nodeID, peerAddrs, peerIDList, func(cmd raft.Command) error {
		if clusterManager != nil {
			return clusterManager.ApplyCommand(cmd)
		}
		return nil
	})

	// Create cluster manager
	clusterManager = raft.NewClusterManager(ctx, raftNode)

	log.Println("Control plane started")

	// Wait for interrupt
	<-ctx.Done()

	log.Println("Shutting down...")

	// Graceful shutdown
	clusterManager.Stop()
	raftNode.Stop()

	select {
	case <-time.After(10 * time.Second):
		log.Println("Forceful shutdown")
	case <-raftNode.Done():
		log.Println("Raft node stopped")
	}

	<-clusterManager.Done()
	log.Println("Control plane stopped")
}

