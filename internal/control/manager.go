package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"seminarska/internal/common/rpc"
	"seminarska/internal/control/dataplane"
	"seminarska/proto/controllink"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
)

type ChainConfig struct {
	LoggerPath      string
	DataExecutable  string
	TargetNodeCount int
}

type ChainManager struct {
	fsm         *ChainFSM
	cfg         ChainConfig
	nodeManager *dataplane.NodeManager
	raft        *raft.Raft
	server      *rpc.Server
	done        chan struct{}
}

func NewChainManager(
	ctx context.Context,
	cfg ChainConfig,
	fsm *ChainFSM,
	raft *raft.Raft,
	addr string,
) *ChainManager {
	m := &ChainManager{
		cfg:         cfg,
		nodeManager: dataplane.NewNodeManager(cfg.DataExecutable),
		done:        make(chan struct{}),
		fsm:         fsm,
		server:      rpc.NewServer(ctx, newClientHandler(fsm), addr),
		raft:        raft,
	}
	go m.init(ctx)
	return m
}

func (m *ChainManager) init(ctx context.Context) {
	log.Println("Starting control plane server")
	go m.SuperviseChain(ctx)
	go m.handleShutdown(ctx)
}

func (m *ChainManager) SuperviseChain(ctx context.Context) {
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			if m.raft.State() == raft.Leader {
				m.runHealthCheck()
			}
		}
	}
}

func (m *ChainManager) runHealthCheck() {
	s := &ChainSnapshot{
		Nodes:   m.fsm.Nodes(),
		Counter: m.fsm.nodeCounter,
	}

	var deadNodes []int
	for i, node := range s.Nodes {
		if err := m.nodeManager.Ping(node); err != nil {
			deadNodes = append(deadNodes, i)
		}
	}
	m.replaceDeadNodes(s, deadNodes)
	m.addMissingNodes(s)
	m.sendStateUpdate(s)
}

func (m *ChainManager) replaceDeadNodes(s *ChainSnapshot, deadNodes []int) {
	if len(deadNodes) == 0 {
		return
	}
	for _, i := range deadNodes {
		log.Println("Node", s.Nodes[i].Config.Id, "is dead")
	}
	m.deleteDeadNodes(s, deadNodes)
	m.rerouteChain(s)
}

func (m *ChainManager) deleteDeadNodes(
	s *ChainSnapshot,
	deadNodes []int,
) {
	remove := make(map[int]struct{}, len(deadNodes))
	for _, i := range deadNodes {
		remove[i] = struct{}{}
	}
	dst := s.Nodes[:0]
	for i, n := range s.Nodes {
		if _, dead := remove[i]; dead {
			_ = m.nodeManager.TerminateDataNode(n)
			continue
		}
		dst = append(dst, n)
	}
	s.Nodes = dst
}

func (m *ChainManager) addNode(s *ChainSnapshot) {
	defer func() { s.Counter++ }()

	node, err := m.spawnNewNode(s)
	if err != nil {
		log.Println("Failed to start new node:", err)
		return
	}

	time.Sleep(time.Millisecond * 500)

	err = m.attachNewNode(s, node)
	if err != nil {
		log.Println("Failed to attach new node:", err)
	}

	log.Println("Added node:", node.Config.Id)

}

func (m *ChainManager) spawnNewNode(s *ChainSnapshot) (*dataplane.NodeDescriptor, error) {
	nextNodeId := fmt.Sprintf("data_%d", s.Counter)
	secret := strconv.Itoa(rand.Int())
	p1, p2, p3 := getNodePorts(s.Counter)
	nodeConfig := dataplane.NewNodeConfig(
		nextNodeId, m.cfg.LoggerPath,
		secret, p1, p2, p3,
	)
	return m.nodeManager.StartNewDataNode(nodeConfig)
}

func (m *ChainManager) attachNewNode(
	s *ChainSnapshot,
	node *dataplane.NodeDescriptor,
) (err error) {
	defer func() {
		if err != nil {
			_ = m.nodeManager.TerminateDataNode(node)
		}
	}()

	if len(s.Nodes) == 0 {
		err = m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageReaderConfirmer)
		if err != nil {
			return
		}
	} else {
		currentTail := s.Nodes[len(s.Nodes)-1]

		if len(s.Nodes) == 1 {
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

	s.Nodes = append(s.Nodes, node)
	return nil
}

func (m *ChainManager) rerouteChain(s *ChainSnapshot) {
	if len(s.Nodes) == 0 {
		log.Fatal("Unrecoverable chain failure: all nodes are dead")
	} else if len(s.Nodes) == 1 {
		err := m.nodeManager.SwitchNodeRole(s.Nodes[0], controllink.NodeRole_MessageReaderConfirmer)
		if err != nil {
			log.Fatal("Failed to recover chain with single node:", err)
		}
		_ = m.nodeManager.DisconnectDataNodeSuccessor(s.Nodes[0])
	} else {
		for i, node := range s.Nodes {
			if i == 0 {
				err := m.nodeManager.SwitchNodeRole(node, controllink.NodeRole_MessageReader)
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
				err = m.nodeManager.SwitchDataNodeSuccessor(node, s.Nodes[i+1])
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
			} else if i == len(s.Nodes)-1 {
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
				err = m.nodeManager.SwitchDataNodeSuccessor(node, s.Nodes[i+1])
				if err != nil {
					log.Fatal("Failed to recover chain with multiple nodes:", err)
				}
			}
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (m *ChainManager) addMissingNodes(s *ChainSnapshot) {
	for range m.cfg.TargetNodeCount - len(s.Nodes) {
		m.addNode(s)
	}
}

func (m *ChainManager) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	m.terminateChain()
	close(m.done)
}

func (m *ChainManager) terminateChain() {
	for _, node := range m.fsm.nodes {
		_ = m.nodeManager.TerminateDataNode(node)
	}
}

func (m *ChainManager) Done() <-chan struct{} {
	return m.done
}

func (m *ChainManager) sendStateUpdate(s *ChainSnapshot) {
	cmd := FullChainCommand{
		Nodes:       s.Nodes,
		NodeCounter: s.Counter,
	}
	data, _ := json.Marshal(cmd)
	f := m.raft.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		log.Println("Failed to apply full-chain command:", err)
		return
	}
}
