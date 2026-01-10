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

type NodeSpawnFunc func() (*ChainNode, error)

type ClusterManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	raftNode      *Node
	healthChecker *HealthChecker

	chainNodes []ChainNode
	chainMu    sync.RWMutex

	spawnNode NodeSpawnFunc

	done chan struct{}
}

func NewClusterManager(ctx context.Context, raftNode *Node) *ClusterManager {
	ctx, cancel := context.WithCancel(ctx)
	cm := &ClusterManager{
		ctx:        ctx,
		cancel:     cancel,
		raftNode:   raftNode,
		chainNodes: make([]ChainNode, 0),
		done:       make(chan struct{}),
	}

	cm.healthChecker = NewHealthChecker(ctx, cm.onNodeFailed)

	go cm.run()
	return cm
}

func (cm *ClusterManager) SetSpawnFunc(fn NodeSpawnFunc) {
	cm.spawnNode = fn
}

func (cm *ClusterManager) AddDataNode(node ChainNode) error {
	if !cm.raftNode.IsLeader() {
		return ErrNotLeader
	}

	payload, _ := json.Marshal(node)
	cmd := Command{
		Type:    CmdAddNode,
		Payload: payload,
	}

	return cm.raftNode.SubmitCommand(cmd)
}

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

func (cm *ClusterManager) onNodeFailed(nodeID string) {
	if !cm.raftNode.IsLeader() {
		return
	}

	log.Printf("[ClusterManager] Node %s failed, reconfiguring chain", nodeID)

	cm.chainMu.RLock()
	chainLen := len(cm.chainNodes)
	var predecessorIdx, successorIdx, failedIdx int = -1, -1, -1

	for i, node := range cm.chainNodes {
		if node.NodeID == nodeID {
			failedIdx = i
			if i > 0 {
				predecessorIdx = i - 1
			}
			if i < chainLen-1 {
				successorIdx = i + 1
			}
			break
		}
	}

	if failedIdx == -1 {
		cm.chainMu.RUnlock()
		return
	}

	isHead := failedIdx == 0
	isTail := failedIdx == chainLen-1

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

	// Rewire predecessor to skip the failed node
	if predecessor != nil {
		newSuccessorAddr := ""
		if successor != nil {
			newSuccessorAddr = successor.ChainAddress
		}
		cm.sendSwitchSuccessor(predecessor.ControlAddress, newSuccessorAddr)
	}

	// Promote new head/tail
	if isHead && successor != nil {
		cm.sendSwitchRole(successor.ControlAddress, controllink.NodeRole_MessageReader)
	}
	if isTail && predecessor != nil {
		cm.sendSwitchRole(predecessor.ControlAddress, controllink.NodeRole_MessageConfirmer)
	}

	if err := cm.RemoveDataNode(nodeID); err != nil {
		log.Printf("[ClusterManager] Failed to remove node %s: %v", nodeID, err)
	}

	go cm.spawnReplacementNode()
}

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
		log.Printf("[ClusterManager] SwitchSuccessor failed for %s: %v", controlAddr, err)
		return
	}

	log.Printf("[ClusterManager] SwitchSuccessor: %s -> %s", controlAddr, newSuccessorAddr)
}

func (cm *ClusterManager) sendSwitchRole(controlAddr string, role controllink.NodeRole) {
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
	_, err = client.SwitchRole(ctx, &controllink.SwitchRoleCommand{Role: role})
	if err != nil {
		log.Printf("[ClusterManager] SwitchRole failed for %s: %v", controlAddr, err)
		return
	}

	log.Printf("[ClusterManager] SwitchRole: %s -> %v", controlAddr, role)
}

func (cm *ClusterManager) spawnReplacementNode() {
	if cm.spawnNode == nil {
		log.Printf("[ClusterManager] No spawn function configured, skipping replacement")
		return
	}

	node, err := cm.spawnNode()
	if err != nil {
		log.Printf("[ClusterManager] Failed to spawn replacement node: %v", err)
		return
	}

	if err := cm.AddDataNode(*node); err != nil {
		log.Printf("[ClusterManager] Failed to add replacement node: %v", err)
		return
	}

	log.Printf("[ClusterManager] Spawned replacement node %s", node.NodeID)
}

func (cm *ClusterManager) run() {
	defer close(cm.done)
	<-cm.ctx.Done()
	cm.healthChecker.Stop()
	<-cm.healthChecker.Done()
}

func (cm *ClusterManager) Stop() {
	cm.cancel()
}

func (cm *ClusterManager) Done() <-chan struct{} {
	return cm.done
}

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

	for _, existing := range cm.chainNodes {
		if existing.NodeID == node.NodeID {
			return ErrAlreadyExists
		}
	}

	node.IsAlive = true
	cm.chainNodes = append(cm.chainNodes, node)
	cm.healthChecker.AddNode(node.NodeID, node.ControlAddress)

	log.Printf("[ClusterManager] Added node %s to chain", node.NodeID)

	// Wire previous tail to new node
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
