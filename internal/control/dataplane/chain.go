//go:build ignore

package dataplane

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
	"strconv"
	"sync"
	"time"
)

const (
	clientPortOffset  = 10_080
	controlPortOffset = 20_080
	dataPortOffset    = 30_080
)

type ChainConfig struct {
	LoggerPath      string
	DataExecutable  string
	TargetNodeCount int
}

type ChainManager struct {
	cfg         ChainConfig
	nodeManager *NodeManager
	nodeCounter int
	nodes       []*NodeDescriptor
	mx          *sync.Mutex
	server      *rpc.Server
	done        chan struct{}
}

func NewChainManager(ctx context.Context, cfg ChainConfig, addr string) *ChainManager {
	m := &ChainManager{
		cfg:         cfg,
		nodeManager: NewNodeManager(cfg.DataExecutable),
		mx:          &sync.Mutex{},
		done:        make(chan struct{}),
		nodeCounter: 0,
	}
	go m.init(ctx, addr)
	return m
}

func (m *ChainManager) Head() *NodeDescriptor {
	m.mx.Lock()
	defer m.mx.Unlock()
	if len(m.nodes) == 0 {
		return nil
	}
	return m.nodes[0]
}

func (m *ChainManager) Mid() *NodeDescriptor {
	m.mx.Lock()
	defer m.mx.Unlock()
	if len(m.nodes) == 0 {
		return nil
	}
	i := rand.Intn(len(m.nodes))
	return m.nodes[i]
}

func (m *ChainManager) Tail() *NodeDescriptor {
	m.mx.Lock()
	defer m.mx.Unlock()
	if len(m.nodes) == 0 {
		return nil
	}
	return m.nodes[len(m.nodes)-1]
}

func (m *ChainManager) init(ctx context.Context, addr string) {
	for range m.cfg.TargetNodeCount {
		m.addNode()
	}
	handler := newClientHandler(m)
	m.server = rpc.NewServer(ctx, handler, addr)
	go m.observeChainHealth(ctx)
	go m.handleShutdown(ctx)
}

func getNodePorts(nodeId int) (string, string, string) {
	return fmt.Sprintf(":%d", clientPortOffset+nodeId),
		fmt.Sprintf(":%d", controlPortOffset+nodeId),
		fmt.Sprintf(":%d", dataPortOffset+nodeId)
}

func (m *ChainManager) addNode() {
	defer func() { m.nodeCounter++ }()

	node, err := m.spawnNewNode()
	if err != nil {
		log.Println("Failed to start new node:", err)
		return
	}

	time.Sleep(time.Millisecond * 500)

	err = m.attachNewNode(node)
	if err != nil {
		log.Println("Failed to attach new node:", err)
	}

	log.Println("Added node:", node.Config.Id)

}

func (m *ChainManager) spawnNewNode() (*NodeDescriptor, error) {
	nextNodeId := fmt.Sprintf("data_%d", m.nodeCounter)
	secret := strconv.Itoa(rand.Int())
	p1, p2, p3 := getNodePorts(m.nodeCounter)
	nodeConfig := NewNodeConfig(
		nextNodeId, m.cfg.LoggerPath,
		secret, p1, p2, p3,
	)
	return m.nodeManager.StartNewDataNode(nodeConfig)
}

func (m *ChainManager) attachNewNode(node *NodeDescriptor) (err error) {
	defer func() {
		if err != nil {
			_ = m.nodeManager.TerminateDataNode(node)
		}
	}()

	if len(m.nodes) == 0 {
		err = m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageReaderConfirmer)
		if err != nil {
			return
		}
	} else {
		currentTail := m.nodes[len(m.nodes)-1]

		if len(m.nodes) == 1 {
			err = m.nodeManager.SwitchNodeRole(currentTail, controllink.NodeRole_MessageReader)
			if err != nil {
				return
			}
		} else {
			err = m.nodeManager.SwitchNodeRole(currentTail, controllink.NodeRole_Relay)
			if err != nil {
				return
			}
		}
		err = m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageConfirmer)
		if err != nil {
			return
		}
		err = m.nodeManager.SwitchDataNodeSuccessor(currentTail, node)
		if err != nil {
			return
		}
	}

	m.nodes = append(m.nodes, node)
	return nil
}

func (m *ChainManager) observeChainHealth(ctx context.Context) {
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			m.runHealthCheck()
		}
	}
}

func (m *ChainManager) runHealthCheck() {
	m.mx.Lock()
	defer m.mx.Unlock()

	var deadNodes []int
	for i, node := range m.nodes {
		if err := m.nodeManager.Ping(node); err != nil {
			deadNodes = append(deadNodes, i)
		}
	}
	m.replaceDeadNodes(deadNodes)
}

func (m *ChainManager) replaceDeadNodes(deadNodes []int) {
	if len(deadNodes) == 0 {
		return
	}
	for _, i := range deadNodes {
		log.Println("Node", m.nodes[i].Config.Id, "is dead")
	}
	m.deleteDeadNodes(deadNodes)
	m.rerouteChain()
	m.addMissingNodes()

}

func (m *ChainManager) deleteDeadNodes(deadNodes []int) {
	remove := make(map[int]struct{}, len(deadNodes))
	for _, i := range deadNodes {
		remove[i] = struct{}{}
	}
	dst := m.nodes[:0]
	for i, n := range m.nodes {
		if _, dead := remove[i]; dead {
			_ = m.nodeManager.TerminateDataNode(n)
			continue
		}
		dst = append(dst, n)
	}
	m.nodes = dst
}

func (m *ChainManager) rerouteChain() {
	if len(m.nodes) == 0 {
		log.Fatal("Unrecoverable chain failure: all nodes are dead")
	} else if len(m.nodes) == 1 {
		err := m.nodeManager.SwitchNodeRole(m.nodes[0], controllink.NodeRole_MessageReaderConfirmer)
		if err != nil {
			log.Fatal("Failed to recover chain with single node:", err)
		}
		_ = m.nodeManager.DisconnectDataNodeSuccessor(m.nodes[0])
	} else {
		for i, node := range m.nodes {
			if i == 0 {
				err := m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageReader)
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
				err = m.nodeManager.SwitchDataNodeSuccessor(node, m.nodes[i+1])
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
			} else if i == len(m.nodes)-1 {
				err := m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageConfirmer)
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
				_ = m.nodeManager.DisconnectDataNodeSuccessor(node)
			} else {
				err := m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_Relay)
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
				err = m.nodeManager.SwitchDataNodeSuccessor(node, m.nodes[i+1])
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
			}
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (m *ChainManager) addMissingNodes() {
	for range m.cfg.TargetNodeCount - len(m.nodes) {
		m.addNode()
	}
}

func (m *ChainManager) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	m.terminateChain()
	close(m.done)
}

func (m *ChainManager) terminateChain() {
	m.mx.Lock()
	defer m.mx.Unlock()
	for _, node := range m.nodes {
		_ = m.nodeManager.TerminateDataNode(node)
	}
}

func (m *ChainManager) Done() <-chan struct{} {
	return m.done
}
