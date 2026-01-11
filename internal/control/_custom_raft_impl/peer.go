package _custom_raft_impl

import (
	"context"
	"log"
	"sync"
	"time"

	raftpb "seminarska/proto/raft"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// peerClient manages RPC communication with a single peer
type peerClient struct {
	ctx     context.Context
	addr    string
	peerID  string
	conn    *grpc.ClientConn
	client  raftpb.RaftServiceClient
	connMu  sync.RWMutex
	healthy bool
}

func newPeerClient(ctx context.Context, addr string, peerID string) *peerClient {
	pc := &peerClient{
		ctx:     ctx,
		addr:    addr,
		peerID:  peerID,
		healthy: false,
	}
	go pc.maintainConnection()
	return pc
}

func (pc *peerClient) maintainConnection() {
	for {
		select {
		case <-pc.ctx.Done():
			pc.closeConnection()
			return
		default:
		}

		if pc.getConn() == nil {
			pc.connect()
		}
		time.Sleep(1 * time.Second)
	}
}

func (pc *peerClient) connect() {
	pc.connMu.Lock()
	defer pc.connMu.Unlock()

	if pc.conn != nil {
		return
	}

	ctx, cancel := context.WithTimeout(pc.ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, pc.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		// Don't spam logs with connection failures
		pc.healthy = false
		return
	}

	pc.conn = conn
	pc.client = raftpb.NewRaftServiceClient(conn)
	pc.healthy = true
	log.Printf("[Raft Peer %s] Connected to %s", pc.peerID, pc.addr)
}

func (pc *peerClient) getConn() *grpc.ClientConn {
	pc.connMu.RLock()
	defer pc.connMu.RUnlock()
	return pc.conn
}

func (pc *peerClient) closeConnection() {
	pc.connMu.Lock()
	defer pc.connMu.Unlock()
	if pc.conn != nil {
		pc.conn.Close()
		pc.conn = nil
		pc.client = nil
	}
}

func (pc *peerClient) getClient() raftpb.RaftServiceClient {
	pc.connMu.RLock()
	defer pc.connMu.RUnlock()
	return pc.client
}

// RequestVote sends a RequestVote RPC to this peer
func (pc *peerClient) RequestVote(req RequestVoteRequest) (RequestVoteResponse, error) {
	client := pc.getClient()
	if client == nil {
		return RequestVoteResponse{}, ErrTimeout
	}

	ctx, cancel := context.WithTimeout(pc.ctx, 150*time.Millisecond)
	defer cancel()

	protoReq := &raftpb.RequestVoteRequest{
		Term:         req.Term,
		CandidateId:  req.CandidateID,
		LastLogIndex: req.LastLogIndex,
		LastLogTerm:  req.LastLogTerm,
	}

	protoResp, err := client.RequestVote(ctx, protoReq)
	if err != nil {
		pc.connMu.Lock()
		pc.healthy = false
		pc.connMu.Unlock()
		return RequestVoteResponse{}, err
	}

	return RequestVoteResponse{
		Term:        protoResp.Term,
		VoteGranted: protoResp.VoteGranted,
	}, nil
}

// AppendEntries sends an AppendEntries RPC to this peer
func (pc *peerClient) AppendEntries(req AppendEntriesRequest) (AppendEntriesResponse, error) {
	client := pc.getClient()
	if client == nil {
		return AppendEntriesResponse{}, ErrTimeout
	}

	ctx, cancel := context.WithTimeout(pc.ctx, 150*time.Millisecond)
	defer cancel()

	// Convert entries to proto format
	protoEntries := make([]*raftpb.LogEntry, len(req.Entries))
	for i, e := range req.Entries {
		protoEntries[i] = &raftpb.LogEntry{
			Term:    e.Term,
			Index:   e.Index,
			Command: e.Command,
		}
	}

	protoReq := &raftpb.AppendEntriesRequest{
		Term:         req.Term,
		LeaderId:     req.LeaderID,
		PrevLogIndex: req.PrevLogIndex,
		PrevLogTerm:  req.PrevLogTerm,
		Entries:      protoEntries,
		LeaderCommit: req.LeaderCommit,
	}

	protoResp, err := client.AppendEntries(ctx, protoReq)
	if err != nil {
		pc.connMu.Lock()
		pc.healthy = false
		pc.connMu.Unlock()
		return AppendEntriesResponse{}, err
	}

	return AppendEntriesResponse{
		Term:    protoResp.Term,
		Success: protoResp.Success,
	}, nil
}

// IsHealthy returns whether this peer is reachable
func (pc *peerClient) IsHealthy() bool {
	pc.connMu.RLock()
	defer pc.connMu.RUnlock()
	return pc.healthy
}
