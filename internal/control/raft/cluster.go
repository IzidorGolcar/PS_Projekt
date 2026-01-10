package raft

import (
	"context"
	"encoding/json"
	"log"
	"seminarska/proto/controllink"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClusterManager manages the data plane cluster state and chain reconfiguration
type ClusterManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	raftNode      *Node
	healthChecker *HealthChecker

	// Current chain state
	chainNodes []ChainNode
	chainMu    sync.RWMutex

	done chan struct{}
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(ctx context.Context, raftNode *Node) *ClusterManager {
	ctx, cancel := context.WithCancel(ctx)
	cm := &ClusterManager{
		ctx:        ctx,
		cancel:     cancel,
		raftNode:   raftNode,
		chainNodes: make([]ChainNode, 0),
		done:       make(chan struct{}),
	}

	// Create health checker with callback
	cm.healthChecker = NewHealthChecker(ctx, cm.onNodeFailed)

	go cm.run()
	return cm
}

// AddDataNode adds a new data node to the chain
func (cm *ClusterManager) AddDataNode(node ChainNode) error {
	if !cm.raftNode.IsLeader() {
		return ErrNotLeader
	}

	// Create command to add node
	payload, _ := json.Marshal(node)
	cmd := Command{
		Type:    CmdAddNode,
		Payload: payload,
	}

	return cm.raftNode.SubmitCommand(cmd)
}

// RemoveDataNode removes a data node from the chain
func (cm *ClusterManager) RemoveDataNode(nodeID string) error {
	if !cm.raftNode.IsLeader() {
		return ErrNotLeader
	}

	payload, _ := json.Marshal(map[string]string{"node_id": nodeID})
	cmd := Command{
		Type:    CmdRemoveNode,
		Payload: payload,
	}

	return cm.raftNode.SubmitCommand(cmd)
}

// GetChainState returns the current chain state
func (cm *ClusterManager) GetChainState() (head, tail *ChainNode, nodes []ChainNode) {
	cm.chainMu.RLock()
	defer cm.chainMu.RUnlock()

	nodes = make([]ChainNode, len(cm.chainNodes))
	copy(nodes, cm.chainNodes)

	if len(nodes) > 0 {
		head = &nodes[0]
		tail = &nodes[len(nodes)-1]
	}

	return
}

// onNodeFailed is called when a data node fails health check
func (cm *ClusterManager) onNodeFailed(nodeID string) {
	if !cm.raftNode.IsLeader() {
		return
	}

	log.Printf("[ClusterManager] Node %s failed, initiating reconfiguration", nodeID)

	cm.chainMu.RLock()
	var predecessorIdx int = -1
	var successorIdx int = -1
	var failedIdx int = -1

	for i, node := range cm.chainNodes {
		if node.NodeID == nodeID {
			failedIdx = i
			if i > 0 {
				predecessorIdx = i - 1
			}
			if i < len(cm.chainNodes)-1 {
				successorIdx = i + 1
			}
			break
		}
	}

	if failedIdx == -1 {
		cm.chainMu.RUnlock()
		return
	}

	var predecessor, successor *ChainNode
	if predecessorIdx >= 0 {
		pred := cm.chainNodes[predecessorIdx]
		predecessor = &pred
	}
	if successorIdx >= 0 {
		succ := cm.chainNodes[successorIdx]
		successor = &succ
	}
	cm.chainMu.RUnlock()

	// If there's a predecessor, tell it to switch to the successor
	if predecessor != nil {
		newSuccessorAddr := ""
		if successor != nil {
			newSuccessorAddr = successor.ChainAddress
		}
		cm.sendSwitchSuccessor(predecessor.ControlAddress, newSuccessorAddr)
	}

	// Submit command to remove the failed node from cluster state
	if err := cm.RemoveDataNode(nodeID); err != nil {
		log.Printf("[ClusterManager] Failed to remove node %s: %v", nodeID, err)
	}
}

// sendSwitchSuccessor sends a SwitchSuccessor command to a data node
func (cm *ClusterManager) sendSwitchSuccessor(controlAddr, newSuccessorAddr string) {
	ctx, cancel := context.WithTimeout(cm.ctx, HealthCheckTimeout*2)
	defer cancel()

	conn, err := grpc.DialContext(ctx, controlAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("[ClusterManager] Failed to connect to %s: %v", controlAddr, err)
		return
	}
	defer conn.Close()

	client := controllink.NewControlServiceClient(conn)
	_, err = client.SwitchSuccessor(ctx, &controllink.SwitchSuccessorCommand{Address: newSuccessorAddr})
	if err != nil {
		log.Printf("[ClusterManager] Failed to send SwitchSuccessor to %s: %v", controlAddr, err)
		return
	}

	log.Printf("[ClusterManager] Sent SwitchSuccessor to %s -> %s", controlAddr, newSuccessorAddr)
}

func (cm *ClusterManager) run() {
	defer close(cm.done)
	<-cm.ctx.Done()
	cm.healthChecker.Stop()
	<-cm.healthChecker.Done()
}

// Stop stops the cluster manager
func (cm *ClusterManager) Stop() {
	cm.cancel()
}

// Done returns a channel that is closed when the cluster manager stops
func (cm *ClusterManager) Done() <-chan struct{} {
	return cm.done
}

// ApplyCommand applies a Raft command to the cluster state
// This is passed as the applyFunc to the Raft node
func (cm *ClusterManager) ApplyCommand(cmd Command) error {
	switch cmd.Type {
	case CmdAddNode:
		var node ChainNode
		if err := json.Unmarshal(cmd.Payload, &node); err != nil {
			return err
		}
		return cm.applyAddNode(node)

	case CmdRemoveNode:
		var data map[string]string
		if err := json.Unmarshal(cmd.Payload, &data); err != nil {
			return err
		}
		return cm.applyRemoveNode(data["node_id"])

	case CmdSetNodeStatus:
		var data struct {
			NodeID  string `json:"node_id"`
			IsAlive bool   `json:"is_alive"`
		}
		if err := json.Unmarshal(cmd.Payload, &data); err != nil {
			return err
		}
		return cm.applySetNodeStatus(data.NodeID, data.IsAlive)

	default:
		log.Printf("[ClusterManager] Unknown command type: %v", cmd.Type)
	}
	return nil
}

func (cm *ClusterManager) applyAddNode(node ChainNode) error {
	cm.chainMu.Lock()
	defer cm.chainMu.Unlock()

	// Check if node already exists
	for _, existing := range cm.chainNodes {
		if existing.NodeID == node.NodeID {
			return ErrAlreadyExists
		}
	}

	// Add to chain
	node.IsAlive = true
	cm.chainNodes = append(cm.chainNodes, node)

	// Add to health checker
	cm.healthChecker.AddNode(node.NodeID, node.ControlAddress)

	log.Printf("[ClusterManager] Added node %s to chain", node.NodeID)

	// If this is not the first node, tell the previous tail to connect to this new node
	if len(cm.chainNodes) > 1 {
		prevTail := cm.chainNodes[len(cm.chainNodes)-2]
		go cm.sendSwitchSuccessor(prevTail.ControlAddress, node.ChainAddress)
	}

	return nil
}

func (cm *ClusterManager) applyRemoveNode(nodeID string) error {
	cm.chainMu.Lock()
	defer cm.chainMu.Unlock()

	newChain := make([]ChainNode, 0, len(cm.chainNodes))
	for _, node := range cm.chainNodes {
		if node.NodeID != nodeID {
			newChain = append(newChain, node)
		}
	}

	if len(newChain) == len(cm.chainNodes) {
		return ErrNodeNotFound
	}

	cm.chainNodes = newChain
	cm.healthChecker.RemoveNode(nodeID)

	log.Printf("[ClusterManager] Removed node %s from chain", nodeID)
	return nil
}

func (cm *ClusterManager) applySetNodeStatus(nodeID string, isAlive bool) error {
	cm.chainMu.Lock()
	defer cm.chainMu.Unlock()

	for i := range cm.chainNodes {
		if cm.chainNodes[i].NodeID == nodeID {
			cm.chainNodes[i].IsAlive = isAlive
			return nil
		}
	}
	return ErrNodeNotFound
}
