package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"seminarska/internal/common/rpc"
	"seminarska/internal/control/dataplane"
	"seminarska/internal/control/raft"
	"seminarska/proto/controllink"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// Parse command line flags
	nodeID := flag.String("id", "", "Node ID (required)")
	addr := flag.String("addr", ":8080", "Address to listen on for client gRPC")
	raftAddr := flag.String("raft-addr", ":7000", "Address for Raft peer-to-peer communication")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (e.g., :7001,:7002)")
	peerIDs := flag.String("peer-ids", "", "Comma-separated list of peer IDs (e.g., node2,node3)")
	dataExec := flag.String("data-exec", "", "Path to data node executable")
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
	log.Printf("Raft address: %s", *raftAddr)
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

	// Create Raft gRPC server for peer-to-peer Raft RPCs (RequestVote, AppendEntries)
	raftServer := raft.NewRaftServer(ctx, *raftAddr, raftNode, clusterManager)
	_ = raftServer // Server runs in background goroutine

	var manager *dataplane.NodeManager
	var nodes []dataplane.NodeDescriptor
	var nodeCounter atomic.Int32
	var nodesMu sync.Mutex

	if *dataExec != "" {
		manager = dataplane.NewNodeManager(*dataExec)
		nodes = launchDataNodes(manager)
		nodeCounter.Store(int32(len(nodes)))
		log.Printf("Launched %d data nodes", len(nodes))

		clusterManager.SetSpawnFunc(func() (*raft.ChainNode, error) {
			nodeNum := nodeCounter.Add(1)
			basePort := 6970 + int(nodeNum)*10

			cfg := dataplane.NewNodeConfig(
				fmt.Sprintf("node%d", nodeNum),
				os.DevNull,
				"secret",
				fmt.Sprintf(":%d", basePort+1),  // control
				fmt.Sprintf(":%d", basePort+11), // chain
				fmt.Sprintf(":%d", basePort+21), // service
			)

			desc, err := manager.StartNewDataNode(cfg)
			if err != nil {
				return nil, err
			}

			time.Sleep(500 * time.Millisecond)

			nodesMu.Lock()
			nodes = append(nodes, *desc)
			nodesMu.Unlock()

			return &raft.ChainNode{
				NodeID:         cfg.Id,
				ControlAddress: cfg.ControlAddress,
				ChainAddress:   cfg.DataChainAddresses,
				ServiceAddress: cfg.ClientRequestsAddress,
				IsAlive:        true,
			}, nil
		})
	}

	// Create and start gRPC server for client requests
	handler := &controlPlaneHandler{
		clusterManager: clusterManager,
		raftNode:       raftNode,
		nodes:          nodes,
	}
	rpcServer := rpc.NewServer(ctx, handler, *addr)

	log.Println("Control plane started")
	log.Printf("Listening on %s", *addr)

	// Wait for interrupt
	<-ctx.Done()

	log.Println("Shutting down...")

	// Terminate data nodes
	if manager != nil {
		for i := range nodes {
			if err := manager.TerminateDataNode(&nodes[i]); err != nil {
				log.Printf("Failed to terminate node: %v", err)
			}
		}
	}

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
	<-rpcServer.Done()
	log.Println("Control plane stopped")
}

// launchDataNodes starts data nodes and connects them in a chain (like mock_control)
func launchDataNodes(manager *dataplane.NodeManager) []dataplane.NodeDescriptor {
	head, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node1", os.DevNull,
		"secret", ":6971", ":6981", ":6991",
	))
	if err != nil {
		log.Printf("Failed to start head node: %v", err)
		return nil
	}

	mid, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node2", os.DevNull,
		"secret", ":6972", ":6982", ":6992",
	))
	if err != nil {
		log.Printf("Failed to start mid node: %v", err)
		return nil
	}

	tail, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node3", os.DevNull,
		"secret", ":6973", ":6983", ":6993",
	))
	if err != nil {
		log.Printf("Failed to start tail node: %v", err)
		return nil
	}

	time.Sleep(time.Second) // Give nodes time to start

	// Set roles
	if err := manager.SwitchNodeRole(head, controllink.NodeRole_MessageReader); err != nil {
		log.Printf("Failed to set head role: %v", err)
	}
	if err := manager.SwitchNodeRole(mid, controllink.NodeRole_Relay); err != nil {
		log.Printf("Failed to set mid role: %v", err)
	}
	if err := manager.SwitchNodeRole(tail, controllink.NodeRole_MessageConfirmer); err != nil {
		log.Printf("Failed to set tail role: %v", err)
	}

	// Connect the chain
	if err := manager.SwitchDataNodeSuccessor(head, mid); err != nil {
		log.Printf("Failed to connect head->mid: %v", err)
	}
	if err := manager.SwitchDataNodeSuccessor(mid, tail); err != nil {
		log.Printf("Failed to connect mid->tail: %v", err)
	}

	return []dataplane.NodeDescriptor{*head, *mid, *tail}
}

// controlPlaneHandler implements the ControlPlane gRPC service
type controlPlaneHandler struct {
	razpravljalnica.UnimplementedControlPlaneServer
	clusterManager *raft.ClusterManager
	raftNode       *raft.Node
	nodes          []dataplane.NodeDescriptor
}

func (h *controlPlaneHandler) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterControlPlaneServer(grpcServer, h)
}

func (h *controlPlaneHandler) GetClusterState(
	_ context.Context, _ *emptypb.Empty,
) (*razpravljalnica.GetClusterStateResponse, error) {
	if len(h.nodes) >= 3 {
		return &razpravljalnica.GetClusterStateResponse{
			Head: h.nodes[0].NodeInfo(),
			Tail: h.nodes[len(h.nodes)-1].NodeInfo(),
		}, nil
	}

	// Fallback to ClusterManager state
	head, tail, _ := h.clusterManager.GetChainState()
	resp := &razpravljalnica.GetClusterStateResponse{}
	if head != nil {
		resp.Head = &razpravljalnica.NodeInfo{
			NodeId:  head.NodeID,
			Address: head.ServiceAddress,
		}
	}
	if tail != nil {
		resp.Tail = &razpravljalnica.NodeInfo{
			NodeId:  tail.NodeID,
			Address: tail.ServiceAddress,
		}
	}
	return resp, nil
}

func (h *controlPlaneHandler) GetSubcscriptionNode(
	_ context.Context, _ *razpravljalnica.SubscriptionNodeRequest,
) (*razpravljalnica.SubscriptionNodeResponse, error) {
	// Return middle node for subscriptions (like mock_control)
	if len(h.nodes) >= 2 {
		midIdx := len(h.nodes) / 2
		return &razpravljalnica.SubscriptionNodeResponse{
			SubscribeToken: h.nodes[midIdx].SubscriptionToken(),
			Node:           h.nodes[midIdx].NodeInfo(),
		}, nil
	}

	// Fallback
	_, _, nodes := h.clusterManager.GetChainState()
	if len(nodes) > 0 {
		midIdx := len(nodes) / 2
		return &razpravljalnica.SubscriptionNodeResponse{
			SubscribeToken: "secret",
			Node: &razpravljalnica.NodeInfo{
				NodeId:  nodes[midIdx].NodeID,
				Address: nodes[midIdx].ServiceAddress,
			},
		}, nil
	}

	return &razpravljalnica.SubscriptionNodeResponse{}, nil
}
