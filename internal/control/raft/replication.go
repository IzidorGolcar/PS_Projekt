package raft

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

var lastHeartbeatTime time.Time

// sendHeartbeats sends AppendEntries RPCs to all peers
// This implements §5.3 of the Raft paper (heartbeats and log replication)
func (n *Node) sendHeartbeats() {
	if !n.IsLeader() {
		return
	}

	// Rate limit heartbeats
	if time.Since(lastHeartbeatTime) < HeartbeatInterval {
		return
	}
	lastHeartbeatTime = time.Now()

	var wg sync.WaitGroup

	for peerID, client := range n.peerClients {
		wg.Add(1)
		go func(peerID string, client *peerClient) {
			defer wg.Done()
			n.sendAppendEntries(peerID, client)
		}(peerID, client)
	}

	// Don't block waiting for all responses
	go func() {
		wg.Wait()
		n.updateCommitIndex()
	}()
}

// sendAppendEntries sends AppendEntries RPC to a single peer
func (n *Node) sendAppendEntries(peerID string, client *peerClient) {
	nextIndex := n.state.GetNextIndex(peerID)
	prevLogIndex := nextIndex - 1
	prevLogTerm := int64(0)

	if prevLogIndex > 0 {
		entry, exists := n.state.GetLogEntry(prevLogIndex)
		if exists {
			prevLogTerm = entry.Term
		}
	}

	// Get entries to send
	entries := n.state.GetEntriesFrom(nextIndex)

	req := AppendEntriesRequest{
		Term:         n.state.GetCurrentTerm(),
		LeaderID:     n.state.GetNodeID(),
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: n.state.GetCommitIndex(),
	}

	resp, err := client.AppendEntries(req)
	if err != nil {
		return
	}

	// If response term is greater, step down
	if resp.Term > n.state.GetCurrentTerm() {
		n.becomeFollower(resp.Term)
		return
	}

	if resp.Success {
		// Update nextIndex and matchIndex for follower
		if len(entries) > 0 {
			lastEntry := entries[len(entries)-1]
			n.state.SetNextIndex(peerID, lastEntry.Index+1)
			n.state.SetMatchIndex(peerID, lastEntry.Index)
		}
	} else {
		// Decrement nextIndex and retry
		if nextIndex > 1 {
			n.state.SetNextIndex(peerID, nextIndex-1)
		}
	}
}

// updateCommitIndex updates the commit index based on matchIndex values
// This implements §5.3 and §5.4 of the Raft paper
func (n *Node) updateCommitIndex() {
	if !n.IsLeader() {
		return
	}

	// Find the highest N such that a majority of matchIndex[i] >= N
	// and log[N].term == currentTerm

	currentTerm := n.state.GetCurrentTerm()
	lastLogIndex := n.state.GetLastLogIndex()

	for N := lastLogIndex; N > n.state.GetCommitIndex(); N-- {
		entry, exists := n.state.GetLogEntry(N)
		if !exists || entry.Term != currentTerm {
			continue
		}

		// Count replicas (including self)
		replicaCount := 1
		for _, peerID := range n.state.GetPeerIDs() {
			if n.state.GetMatchIndex(peerID) >= N {
				replicaCount++
			}
		}

		// Check for majority
		totalNodes := len(n.state.GetPeers()) + 1
		if replicaCount > totalNodes/2 {
			n.state.SetCommitIndex(N)
			n.applyCommittedEntries()
			break
		}
	}
}

// applyCommittedEntries applies all committed but not yet applied entries
func (n *Node) applyCommittedEntries() {
	n.state.mu.Lock()
	commitIndex := n.state.commitIndex
	lastApplied := n.state.lastApplied
	n.state.mu.Unlock()

	for i := lastApplied + 1; i <= commitIndex; i++ {
		entry, exists := n.state.GetLogEntry(i)
		if !exists {
			continue
		}

		// Apply the entry
		n.applyEntry(entry)

		n.state.mu.Lock()
		n.state.lastApplied = i
		n.state.mu.Unlock()
	}
}

// applyEntry applies a single log entry to the state machine
func (n *Node) applyEntry(entry LogEntry) {
	var cmd Command
	if err := json.Unmarshal(entry.Command, &cmd); err != nil {
		log.Printf("[Raft %s] Failed to unmarshal command: %v", n.state.GetNodeID(), err)
		return
	}

	if n.applyFunc != nil {
		if err := n.applyFunc(cmd); err != nil {
			log.Printf("[Raft %s] Failed to apply command: %v", n.state.GetNodeID(), err)
		}
	}

	log.Printf("[Raft %s] Applied entry %d", n.state.GetNodeID(), entry.Index)
}
