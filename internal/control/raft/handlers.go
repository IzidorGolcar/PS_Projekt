package raft

import (
	"encoding/json"
	"log"
)

// handleRequestVote processes incoming RequestVote RPCs
// This implements §5.2 of the Raft paper
func (n *Node) handleRequestVote(req RequestVoteRequest) RequestVoteResponse {
	currentTerm := n.state.GetCurrentTerm()

	// Reply false if term < currentTerm (§5.1)
	if req.Term < currentTerm {
		return RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}
	}

	// If RPC request contains term T > currentTerm: set currentTerm = T, convert to follower (§5.1)
	if req.Term > currentTerm {
		n.becomeFollower(req.Term)
		currentTerm = req.Term
	}

	votedFor := n.state.GetVotedFor()
	lastLogIndex := n.state.GetLastLogIndex()
	lastLogTerm := n.state.GetLastLogTerm()

	// If votedFor is null or candidateId, and candidate's log is at least as up-to-date as receiver's log, grant vote (§5.2, §5.4)
	logIsUpToDate := req.LastLogTerm > lastLogTerm ||
		(req.LastLogTerm == lastLogTerm && req.LastLogIndex >= lastLogIndex)

	voteGranted := false
	if (votedFor == "" || votedFor == req.CandidateID) && logIsUpToDate {
		n.state.SetVotedFor(req.CandidateID)
		n.state.ResetElectionTimeout()
		voteGranted = true
		log.Printf("[Raft %s] Voted for %s in term %d", n.state.GetNodeID(), req.CandidateID, req.Term)
	}

	return RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: voteGranted,
	}
}

// handleAppendEntries processes incoming AppendEntries RPCs
// This implements §5.3 of the Raft paper
func (n *Node) handleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	currentTerm := n.state.GetCurrentTerm()

	// Reply false if term < currentTerm (§5.1)
	if req.Term < currentTerm {
		return AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}
	}

	// If RPC request contains term T > currentTerm: set currentTerm = T, convert to follower (§5.1)
	if req.Term > currentTerm {
		n.becomeFollower(req.Term)
		currentTerm = req.Term
	}

	// Reset election timeout - we've heard from the leader
	n.state.ResetElectionTimeout()
	n.state.SetLeaderID(req.LeaderID)

	// If we're a candidate and receive AppendEntries from legitimate leader, become follower
	if n.state.GetState() == Candidate {
		n.state.SetState(Follower)
	}

	// Reply false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm (§5.3)
	if req.PrevLogIndex > 0 {
		entry, exists := n.state.GetLogEntry(req.PrevLogIndex)
		if !exists || entry.Term != req.PrevLogTerm {
			return AppendEntriesResponse{
				Term:    currentTerm,
				Success: false,
			}
		}
	}

	// If an existing entry conflicts with a new one (same index but different terms),
	// delete the existing entry and all that follow it (§5.3)
	for _, newEntry := range req.Entries {
		existingEntry, exists := n.state.GetLogEntry(newEntry.Index)
		if exists && existingEntry.Term != newEntry.Term {
			n.state.DeleteEntriesFrom(newEntry.Index)
			break
		}
	}

	// Append any new entries not already in the log
	for _, newEntry := range req.Entries {
		_, exists := n.state.GetLogEntry(newEntry.Index)
		if !exists {
			n.state.AppendEntry(newEntry)
		}
	}

	// If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if req.LeaderCommit > n.state.GetCommitIndex() {
		lastNewIndex := n.state.GetLastLogIndex()
		newCommitIndex := req.LeaderCommit
		if lastNewIndex < newCommitIndex {
			newCommitIndex = lastNewIndex
		}
		n.state.SetCommitIndex(newCommitIndex)

		// Apply newly committed entries
		n.applyCommittedEntries()
	}

	return AppendEntriesResponse{
		Term:    currentTerm,
		Success: true,
	}
}

// handleCommand handles a new command submission (only on leader)
func (n *Node) handleCommand(cmd Command) error {
	if !n.IsLeader() {
		return ErrNotLeader
	}

	// Serialize the command
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// Create new log entry
	newIndex := n.state.GetLastLogIndex() + 1
	entry := LogEntry{
		Term:    n.state.GetCurrentTerm(),
		Index:   newIndex,
		Command: cmdBytes,
	}

	// Append to local log
	n.state.AppendEntry(entry)

	log.Printf("[Raft %s] Appended entry %d to log", n.state.GetNodeID(), newIndex)

	// Will be replicated to followers via heartbeat/AppendEntries
	return nil
}

// IsLeader returns true if this node is the current leader
func (n *Node) IsLeader() bool {
	return n.state.GetState() == Leader
}
