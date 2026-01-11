package control

import (
	"encoding/json"
	"io"
	"math/rand"
	"seminarska/internal/control/dataplane"
	"sync"

	"github.com/hashicorp/raft"
)

type FullChainCommand struct {
	Nodes       []*dataplane.NodeDescriptor `json:"nodes"`
	NodeCounter int                         `json:"node_counter"`
}

type ChainFSM struct {
	mx          sync.Mutex
	nodes       []*dataplane.NodeDescriptor
	nodeCounter int
}

func NewChainFSM() *ChainFSM {
	return &ChainFSM{
		nodes: []*dataplane.NodeDescriptor{},
	}
}

func (c *ChainFSM) Apply(log *raft.Log) any {
	var cmd FullChainCommand
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	c.mx.Lock()
	defer c.mx.Unlock()
	c.nodes = cmd.Nodes
	c.nodeCounter = cmd.NodeCounter
	return nil
}

func (c *ChainFSM) Snapshot() (raft.FSMSnapshot, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	nodesCopy := make([]*dataplane.NodeDescriptor, len(c.nodes))
	copy(nodesCopy, c.nodes)
	return &ChainSnapshot{
		Nodes:   nodesCopy,
		Counter: c.nodeCounter,
	}, nil
}

func (c *ChainFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var snap ChainSnapshot
	if err := json.NewDecoder(rc).Decode(&snap); err != nil {
		return err
	}
	c.mx.Lock()
	defer c.mx.Unlock()
	c.nodes = snap.Nodes
	c.nodeCounter = snap.Counter
	return nil
}

func (c *ChainFSM) Nodes() []*dataplane.NodeDescriptor {
	c.mx.Lock()
	defer c.mx.Unlock()
	nodesCopy := make([]*dataplane.NodeDescriptor, len(c.nodes))
	copy(nodesCopy, c.nodes)
	return nodesCopy
}

type ChainSnapshot struct {
	Nodes   []*dataplane.NodeDescriptor `json:"nodes"`
	Counter int                         `json:"counter"`
}

func (s *ChainSnapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	return json.NewEncoder(sink).Encode(s)
}

func (s *ChainSnapshot) Release() {}

func (c *ChainFSM) Head() *dataplane.NodeDescriptor {
	c.mx.Lock()
	defer c.mx.Unlock()
	if len(c.nodes) == 0 {
		return nil
	}
	return c.nodes[0]
}

func (c *ChainFSM) Mid() *dataplane.NodeDescriptor {
	c.mx.Lock()
	defer c.mx.Unlock()
	if len(c.nodes) == 0 {
		return nil
	}
	i := rand.Intn(len(c.nodes))
	return c.nodes[i]
}

func (c *ChainFSM) Tail() *dataplane.NodeDescriptor {
	c.mx.Lock()
	defer c.mx.Unlock()
	if len(c.nodes) == 0 {
		return nil
	}
	return c.nodes[len(c.nodes)-1]
}
