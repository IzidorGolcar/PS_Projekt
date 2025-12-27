package raft

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// peerClient manages RPC communication with a single peer
type peerClient struct {
	ctx     context.Context
	addr    string
	peerID  string
	conn    *grpc.ClientConn
	connMu  sync.RWMutex
	healthy bool
}

func newPeerClient(ctx context.Context, addr string, peerID string) *peerClient {
	pc := &peerClient{
		ctx:     ctx,
		addr:    addr,
		peerID:  peerID,
		healthy: true,
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
		log.Printf("[Raft Peer %s] Failed to connect to %s: %v", pc.peerID, pc.addr, err)
		pc.healthy = false
		return
	}

	pc.conn = conn
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
	}
}

// RequestVote sends a RequestVote RPC to this peer
func (pc *peerClient) RequestVote(req RequestVoteRequest) (RequestVoteResponse, error) {
	conn := pc.getConn()
	if conn == nil {
		return RequestVoteResponse{}, ErrTimeout
	}

	ctx, cancel := context.WithTimeout(pc.ctx, 100*time.Millisecond)
	defer cancel()

	// For now, we'll use a simple unary RPC
	// This will be replaced with the generated proto client
	var resp RequestVoteResponse

	// TODO: Replace with actual gRPC call when proto is generated
	// client := raft.NewRaftServiceClient(conn)
	// protoResp, err := client.RequestVote(ctx, &raft.RequestVoteRequest{...})

	// Simulate the RPC for now - will be replaced with real implementation
	_ = ctx
	resp = RequestVoteResponse{
		Term:        req.Term,
		VoteGranted: false,
	}

	return resp, nil
}

// AppendEntries sends an AppendEntries RPC to this peer
func (pc *peerClient) AppendEntries(req AppendEntriesRequest) (AppendEntriesResponse, error) {
	conn := pc.getConn()
	if conn == nil {
		return AppendEntriesResponse{}, ErrTimeout
	}

	ctx, cancel := context.WithTimeout(pc.ctx, 100*time.Millisecond)
	defer cancel()

	// TODO: Replace with actual gRPC call when proto is generated
	_ = ctx
	resp := AppendEntriesResponse{
		Term:    req.Term,
		Success: true,
	}

	return resp, nil
}

// IsHealthy returns whether this peer is reachable
func (pc *peerClient) IsHealthy() bool {
	pc.connMu.RLock()
	defer pc.connMu.RUnlock()
	return pc.healthy
}
