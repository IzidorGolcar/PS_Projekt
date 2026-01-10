package raft

import (
	"context"
	"log"
	"net"

	raftpb "seminarska/proto/raft"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RaftServer implements the RaftService gRPC server
type RaftServer struct {
	raftpb.UnimplementedRaftServiceServer
	node *Node
	cm   *ClusterManager

	grpcServer *grpc.Server
	addr       string
	done       chan struct{}
}

// NewRaftServer creates and starts a new Raft gRPC server
func NewRaftServer(ctx context.Context, addr string, node *Node, cm *ClusterManager) *RaftServer {
	s := &RaftServer{
		node:       node,
		cm:         cm,
		grpcServer: grpc.NewServer(),
		addr:       addr,
		done:       make(chan struct{}),
	}

	raftpb.RegisterRaftServiceServer(s.grpcServer, s)

	go s.serve()
	go s.handleShutdown(ctx)

	return s
}

func (s *RaftServer) serve() {
	defer close(s.done)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("[RaftServer] Failed to listen on %s: %v", s.addr, err)
	}

	log.Printf("[RaftServer] Listening on %s", s.addr)
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Printf("[RaftServer] Server stopped: %v", err)
	}
}

func (s *RaftServer) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	s.grpcServer.GracefulStop()
}

// Done returns a channel that closes when the server stops
func (s *RaftServer) Done() <-chan struct{} {
	return s.done
}

// RequestVote handles incoming RequestVote RPCs from candidates
func (s *RaftServer) RequestVote(ctx context.Context, req *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {
	// Convert proto to internal type
	internalReq := RequestVoteRequest{
		Term:         req.Term,
		CandidateID:  req.CandidateId,
		LastLogIndex: req.LastLogIndex,
		LastLogTerm:  req.LastLogTerm,
	}

	// Send to node's channel and wait for response
	respChan := make(chan RequestVoteResponse, 1)
	s.node.requestVoteCh <- requestVoteReq{
		req:      internalReq,
		respChan: respChan,
	}

	select {
	case resp := <-respChan:
		return &raftpb.RequestVoteResponse{
			Term:        resp.Term,
			VoteGranted: resp.VoteGranted,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// AppendEntries handles incoming AppendEntries RPCs from leader
func (s *RaftServer) AppendEntries(ctx context.Context, req *raftpb.AppendEntriesRequest) (*raftpb.AppendEntriesResponse, error) {
	// Convert proto entries to internal type
	entries := make([]LogEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = LogEntry{
			Term:    e.Term,
			Index:   e.Index,
			Command: e.Command,
		}
	}

	internalReq := AppendEntriesRequest{
		Term:         req.Term,
		LeaderID:     req.LeaderId,
		PrevLogIndex: req.PrevLogIndex,
		PrevLogTerm:  req.PrevLogTerm,
		Entries:      entries,
		LeaderCommit: req.LeaderCommit,
	}

	// Send to node's channel and wait for response
	respChan := make(chan AppendEntriesResponse, 1)
	s.node.appendEntriesCh <- appendEntriesReq{
		req:      internalReq,
		respChan: respChan,
	}

	select {
	case resp := <-respChan:
		return &raftpb.AppendEntriesResponse{
			Term:    resp.Term,
			Success: resp.Success,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RegisterNode handles requests to add a new data node to the cluster
func (s *RaftServer) RegisterNode(ctx context.Context, req *raftpb.RegisterNodeRequest) (*raftpb.RegisterNodeResponse, error) {
	if !s.node.IsLeader() {
		return &raftpb.RegisterNodeResponse{
			Success:    false,
			LeaderHint: s.node.state.GetLeaderID(),
		}, nil
	}

	node := ChainNode{
		NodeID:         req.Node.NodeId,
		ServiceAddress: req.Node.ServiceAddress,
		ChainAddress:   req.Node.ChainAddress,
		ControlAddress: req.Node.ControlAddress,
		IsAlive:        true,
	}

	err := s.cm.AddDataNode(node)
	if err != nil {
		return &raftpb.RegisterNodeResponse{Success: false}, nil
	}

	return &raftpb.RegisterNodeResponse{Success: true}, nil
}

// UnregisterNode handles requests to remove a data node from the cluster
func (s *RaftServer) UnregisterNode(ctx context.Context, req *raftpb.UnregisterNodeRequest) (*emptypb.Empty, error) {
	if s.node.IsLeader() {
		s.cm.RemoveDataNode(req.NodeId)
	}
	return &emptypb.Empty{}, nil
}

// GetClusterInfo returns the current cluster state
func (s *RaftServer) GetClusterInfo(ctx context.Context, _ *emptypb.Empty) (*raftpb.ClusterInfoResponse, error) {
	head, tail, nodes := s.cm.GetChainState()

	protoNodes := make([]*raftpb.ChainNode, len(nodes))
	for i, n := range nodes {
		protoNodes[i] = &raftpb.ChainNode{
			NodeId:         n.NodeID,
			ServiceAddress: n.ServiceAddress,
			ChainAddress:   n.ChainAddress,
			ControlAddress: n.ControlAddress,
			IsAlive:        n.IsAlive,
		}
	}

	resp := &raftpb.ClusterInfoResponse{
		State: &raftpb.ClusterState{
			Nodes: protoNodes,
		},
		LeaderId: s.node.state.GetLeaderID(),
		Term:     s.node.state.GetCurrentTerm(),
	}

	if head != nil {
		resp.State.HeadId = head.NodeID
	}
	if tail != nil {
		resp.State.TailId = tail.NodeID
	}

	return resp, nil
}

