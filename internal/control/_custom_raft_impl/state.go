package _custom_raft_impl

import (
	"sync"
	"time"
)

// NodeState represents the state of a Raft node
type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (s NodeState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// LogEntry represents a single entry in the Raft log
type LogEntry struct {
	Term    int64
	Index   int64
	Command []byte
}

// RaftState holds the persistent and volatile state of a Raft node
type RaftState struct {
	mu sync.RWMutex

	// Persistent state (should be persisted to stable storage)
	currentTerm int64
	votedFor    string // candidateId that received vote in current term (empty if none)
	log         []LogEntry

	// Volatile state on all servers
	commitIndex int64 // index of highest log entry known to be committed
	lastApplied int64 // index of highest log entry applied to state machine

	// Volatile state on leaders (reinitialized after election)
	nextIndex  map[string]int64 // for each server, index of next log entry to send
	matchIndex map[string]int64 // for each server, index of highest log entry known to be replicated

	// Node identity and cluster configuration
	nodeID   string
	state    NodeState
	leaderID string
	peers    []string // addresses of other nodes in the cluster
	peerIDs  []string // IDs of other nodes

	// Timing
	lastHeartbeat   time.Time
	electionTimeout time.Duration
}

// NewRaftState creates a new RaftState instance
func NewRaftState(nodeID string, peers []string, peerIDs []string) *RaftState {
	return &RaftState{
		currentTerm:     0,
		votedFor:        "",
		log:             make([]LogEntry, 0),
		commitIndex:     0,
		lastApplied:     0,
		nextIndex:       make(map[string]int64),
		matchIndex:      make(map[string]int64),
		nodeID:          nodeID,
		state:           Follower,
		leaderID:        "",
		peers:           peers,
		peerIDs:         peerIDs,
		lastHeartbeat:   time.Now(),
		electionTimeout: randomElectionTimeout(),
	}
}

// Getters with proper locking
func (r *RaftState) GetCurrentTerm() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentTerm
}

func (r *RaftState) GetState() NodeState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *RaftState) GetLeaderID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.leaderID
}

func (r *RaftState) GetNodeID() string {
	return r.nodeID // immutable, no lock needed
}

func (r *RaftState) GetVotedFor() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.votedFor
}

func (r *RaftState) GetLastLogIndex() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.log) == 0 {
		return 0
	}
	return r.log[len(r.log)-1].Index
}

func (r *RaftState) GetLastLogTerm() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.log) == 0 {
		return 0
	}
	return r.log[len(r.log)-1].Term
}

func (r *RaftState) GetCommitIndex() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commitIndex
}

// Setters with proper locking

func (r *RaftState) SetTerm(term int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentTerm = term
	r.votedFor = "" // reset vote when term changes
}

func (r *RaftState) SetVotedFor(candidateID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.votedFor = candidateID
}

func (r *RaftState) SetState(state NodeState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = state
}

func (r *RaftState) SetLeaderID(leaderID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leaderID = leaderID
}

func (r *RaftState) SetCommitIndex(index int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if index > r.commitIndex {
		r.commitIndex = index
	}
}

func (r *RaftState) UpdateLastHeartbeat() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastHeartbeat = time.Now()
}

func (r *RaftState) ResetElectionTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.electionTimeout = randomElectionTimeout()
	r.lastHeartbeat = time.Now()
}

func (r *RaftState) IsElectionTimeoutElapsed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return time.Since(r.lastHeartbeat) > r.electionTimeout
}

// Log operations

func (r *RaftState) AppendEntry(entry LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log = append(r.log, entry)
}

func (r *RaftState) GetLogEntry(index int64) (LogEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, entry := range r.log {
		if entry.Index == index {
			return entry, true
		}
	}
	return LogEntry{}, false
}

func (r *RaftState) GetEntriesFrom(index int64) []LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := make([]LogEntry, 0)
	for _, entry := range r.log {
		if entry.Index >= index {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (r *RaftState) DeleteEntriesFrom(index int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	newLog := make([]LogEntry, 0)
	for _, entry := range r.log {
		if entry.Index < index {
			newLog = append(newLog, entry)
		}
	}
	r.log = newLog
}

// Leader state management

func (r *RaftState) InitLeaderState() {
	r.mu.Lock()
	defer r.mu.Unlock()
	lastLogIndex := int64(0)
	if len(r.log) > 0 {
		lastLogIndex = r.log[len(r.log)-1].Index
	}
	for _, peerID := range r.peerIDs {
		r.nextIndex[peerID] = lastLogIndex + 1
		r.matchIndex[peerID] = 0
	}
}

func (r *RaftState) GetNextIndex(peerID string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nextIndex[peerID]
}

func (r *RaftState) SetNextIndex(peerID string, index int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextIndex[peerID] = index
}

func (r *RaftState) GetMatchIndex(peerID string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.matchIndex[peerID]
}

func (r *RaftState) SetMatchIndex(peerID string, index int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matchIndex[peerID] = index
}
func (r *RaftState) GetPeers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.peers
}

func (r *RaftState) GetPeerIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.peerIDs
}
