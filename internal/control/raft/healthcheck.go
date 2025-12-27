package raft

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HealthChecker monitors the health of data nodes in the chain
type HealthChecker struct {
	ctx    context.Context
	cancel context.CancelFunc

	// Callback when a node fails
	onNodeFailed func(nodeID string)

	// Data nodes to monitor
	nodes   map[string]*dataNodeClient
	nodesMu sync.RWMutex

	done chan struct{}
}

// dataNodeClient manages connection to a data node for health checks
type dataNodeClient struct {
	nodeID         string
	controlAddress string
	conn           *grpc.ClientConn
	healthy        bool
	failCount      int
	mu             sync.Mutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(ctx context.Context, onNodeFailed func(nodeID string)) *HealthChecker {
	ctx, cancel := context.WithCancel(ctx)
	hc := &HealthChecker{
		ctx:          ctx,
		cancel:       cancel,
		onNodeFailed: onNodeFailed,
		nodes:        make(map[string]*dataNodeClient),
		done:         make(chan struct{}),
	}
	go hc.run()
	return hc
}

// AddNode adds a data node to be monitored
func (hc *HealthChecker) AddNode(nodeID, controlAddress string) {
	hc.nodesMu.Lock()
	defer hc.nodesMu.Unlock()

	if _, exists := hc.nodes[nodeID]; exists {
		return
	}

	hc.nodes[nodeID] = &dataNodeClient{
		nodeID:         nodeID,
		controlAddress: controlAddress,
		healthy:        true,
		failCount:      0,
	}
	log.Printf("[HealthChecker] Added node %s at %s", nodeID, controlAddress)
}

// RemoveNode removes a data node from monitoring
func (hc *HealthChecker) RemoveNode(nodeID string) {
	hc.nodesMu.Lock()
	defer hc.nodesMu.Unlock()

	if client, exists := hc.nodes[nodeID]; exists {
		if client.conn != nil {
			client.conn.Close()
		}
		delete(hc.nodes, nodeID)
		log.Printf("[HealthChecker] Removed node %s", nodeID)
	}
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.cancel()
}

// Done returns a channel that is closed when the health checker stops
func (hc *HealthChecker) Done() <-chan struct{} {
	return hc.done
}

func (hc *HealthChecker) run() {
	defer close(hc.done)

	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.checkAllNodes()
		}
	}
}

func (hc *HealthChecker) checkAllNodes() {
	hc.nodesMu.RLock()
	nodes := make([]*dataNodeClient, 0, len(hc.nodes))
	for _, node := range hc.nodes {
		nodes = append(nodes, node)
	}
	hc.nodesMu.RUnlock()

	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(node *dataNodeClient) {
			defer wg.Done()
			hc.checkNode(node)
		}(node)
	}
	wg.Wait()
}

func (hc *HealthChecker) checkNode(node *dataNodeClient) {
	node.mu.Lock()
	defer node.mu.Unlock()

	// Try to connect if no connection
	if node.conn == nil {
		ctx, cancel := context.WithTimeout(hc.ctx, HealthCheckTimeout)
		conn, err := grpc.DialContext(ctx, node.controlAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()

		if err != nil {
			node.failCount++
			if node.failCount >= 3 && node.healthy {
				node.healthy = false
				log.Printf("[HealthChecker] Node %s is DOWN after %d failures", node.nodeID, node.failCount)
				if hc.onNodeFailed != nil {
					go hc.onNodeFailed(node.nodeID)
				}
			}
			return
		}
		node.conn = conn
	}

	// Connection exists - node is healthy
	node.failCount = 0
	if !node.healthy {
		node.healthy = true
		log.Printf("[HealthChecker] Node %s is UP", node.nodeID)
	}
}
