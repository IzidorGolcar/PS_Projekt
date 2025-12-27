package raft

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Command represents a state machine command
type Command struct {
	Type    CommandType
	Payload json.RawMessage
}

type CommandType int

const (
	CmdAddNode CommandType = iota
	CmdRemoveNode
	CmdSetNodeStatus
	CmdReconfigureChain
)

// ChainNode represents a node in the data plane chain
type ChainNode struct {
	NodeID         string
	ServiceAddress string
	ChainAddress   string
	ControlAddress string
	IsAlive        bool
}

// ClusterState represents the current state of the data plane cluster
type ClusterState struct {
	Nodes  []ChainNode
	HeadID string
	TailID string
}

// ApplyFunc is called when a command is committed and should be applied
type ApplyFunc func(cmd Command) error

// Node represents a Raft consensus node
type Node struct {
	ctx    context.Context
	cancel context.CancelFunc

	state *RaftState

	// Cluster state (the state machine)
	clusterState ClusterState
	clusterMu    sync.RWMutex

	// Callback for applying committed commands
	applyFunc ApplyFunc

	// Channels for internal communication
	appendEntriesCh chan appendEntriesReq
	requestVoteCh   chan requestVoteReq
	commandCh       chan commandReq
	commitCh        chan LogEntry

	// RPC clients to peers
	peerClients map[string]*peerClient

	done chan struct{}
}

// Request/response types for internal channels
type appendEntriesReq struct {
	req      AppendEntriesRequest
	respChan chan AppendEntriesResponse
}

type requestVoteReq struct {
	req      RequestVoteRequest
	respChan chan RequestVoteResponse
}

type commandReq struct {
	cmd      Command
	respChan chan error
}

// RPC request/response types (simplified, will be replaced by proto generated types)
type AppendEntriesRequest struct {
	Term         int64
	LeaderID     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []LogEntry
	LeaderCommit int64
}

type AppendEntriesResponse struct {
	Term    int64
	Success bool
}

type RequestVoteRequest struct {
	Term         int64
	CandidateID  string
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteResponse struct {
	Term        int64
	VoteGranted bool
}

// NewNode creates a new Raft node
func NewNode(ctx context.Context, nodeID string, peers []string, peerIDs []string, applyFunc ApplyFunc) *Node {
	ctx, cancel := context.WithCancel(ctx)

	n := &Node{
		ctx:             ctx,
		cancel:          cancel,
		state:           NewRaftState(nodeID, peers, peerIDs),
		clusterState:    ClusterState{Nodes: make([]ChainNode, 0)},
		applyFunc:       applyFunc,
		appendEntriesCh: make(chan appendEntriesReq, 100),
		requestVoteCh:   make(chan requestVoteReq, 100),
		commandCh:       make(chan commandReq, 100),
		commitCh:        make(chan LogEntry, 100),
		peerClients:     make(map[string]*peerClient),
		done:            make(chan struct{}),
	}

	// Initialize peer clients
	for i, addr := range peers {
		n.peerClients[peerIDs[i]] = newPeerClient(ctx, addr, peerIDs[i])
	}

	go n.run()
	return n
}

// Done returns a channel that is closed when the node stops
func (n *Node) Done() <-chan struct{} {
	return n.done
}

// Stop gracefully stops the Raft node
func (n *Node) Stop() {
	n.cancel()
}

// SubmitCommand submits a command to be replicated (only works on leader)
func (n *Node) SubmitCommand(cmd Command) error {
	if !n.IsLeader() {
		return ErrNotLeader
	}
	respChan := make(chan error, 1)
	n.commandCh <- commandReq{cmd: cmd, respChan: respChan}
	return <-respChan
}

// run is the main loop of the Raft node
func (n *Node) run() {
	defer close(n.done)
	log.Printf("[Raft %s] Starting as %s", n.state.GetNodeID(), n.state.GetState())

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			log.Printf("[Raft %s] Shutting down", n.state.GetNodeID())
			return

		case <-ticker.C:
			n.tick()

		case req := <-n.appendEntriesCh:
			resp := n.handleAppendEntries(req.req)
			req.respChan <- resp

		case req := <-n.requestVoteCh:
			resp := n.handleRequestVote(req.req)
			req.respChan <- resp

		case req := <-n.commandCh:
			err := n.handleCommand(req.cmd)
			req.respChan <- err

		case entry := <-n.commitCh:
			n.applyEntry(entry)
		}
	}
}

// tick handles time-based events (election timeout, heartbeats)
func (n *Node) tick() {
	switch n.state.GetState() {
	case Follower:
		if n.state.IsElectionTimeoutElapsed() {
			n.startElection()
		}
	case Candidate:
		if n.state.IsElectionTimeoutElapsed() {
			n.startElection() // start new election
		}
	case Leader:
		n.sendHeartbeats()
	}
}

// startElection transitions to candidate and starts an election
func (n *Node) startElection() {
	n.state.mu.Lock()
	n.state.currentTerm++
	n.state.state = Candidate
	n.state.votedFor = n.state.nodeID
	n.state.electionTimeout = randomElectionTimeout()
	n.state.lastHeartbeat = time.Now()
	currentTerm := n.state.currentTerm
	lastLogIndex := int64(0)
	lastLogTerm := int64(0)
	if len(n.state.log) > 0 {
		lastLogIndex = n.state.log[len(n.state.log)-1].Index
		lastLogTerm = n.state.log[len(n.state.log)-1].Term
	}
	n.state.mu.Unlock()

	log.Printf("[Raft %s] Starting election for term %d", n.state.GetNodeID(), currentTerm)

	// Request votes from all peers
	votesReceived := 1 // vote for self
	votesNeeded := (len(n.state.GetPeers())+1)/2 + 1

	var wg sync.WaitGroup
	var votesMu sync.Mutex

	for peerID, client := range n.peerClients {
		wg.Add(1)
		go func(peerID string, client *peerClient) {
			defer wg.Done()

			resp, err := client.RequestVote(RequestVoteRequest{
				Term:         currentTerm,
				CandidateID:  n.state.GetNodeID(),
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			})

			if err != nil {
				log.Printf("[Raft %s] RequestVote to %s failed: %v", n.state.GetNodeID(), peerID, err)
				return
			}

			if resp.Term > currentTerm {
				n.becomeFollower(resp.Term)
				return
			}

			if resp.VoteGranted {
				votesMu.Lock()
				votesReceived++
				votesMu.Unlock()
			}
		}(peerID, client)
	}

	// Wait for responses with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	case <-n.ctx.Done():
		return
	}

	// Check if we won
	votesMu.Lock()
	votes := votesReceived
	votesMu.Unlock()

	if n.state.GetState() != Candidate {
		return // already transitioned
	}

	if votes >= votesNeeded {
		n.becomeLeader()
	}
}

// becomeFollower transitions to follower state
func (n *Node) becomeFollower(term int64) {
	log.Printf("[Raft %s] Becoming follower for term %d", n.state.GetNodeID(), term)
	n.state.SetTerm(term)
	n.state.SetState(Follower)
	n.state.ResetElectionTimeout()
}

// becomeLeader transitions to leader state
func (n *Node) becomeLeader() {
	log.Printf("[Raft %s] Becoming leader for term %d", n.state.GetNodeID(), n.state.GetCurrentTerm())
	n.state.SetState(Leader)
	n.state.SetLeaderID(n.state.GetNodeID())
	n.state.InitLeaderState()

	// Send initial heartbeats
	n.sendHeartbeats()
}
