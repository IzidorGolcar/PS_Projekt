package chain

import (
	"context"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/internal/common/stream"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	l         *listener
	rpcServer *rpc.Server
}

func NewServer(ctx context.Context, state *nodeDFA, addr string, buffer int) *Server {
	l := newListener(state, buffer)
	return &Server{
		l:         l,
		rpcServer: rpc.NewServer(ctx, l, addr),
	}
}

func (s *Server) Outbound() chan<- *datalink.Confirmation {
	return s.l.outbound
}

func (s *Server) Inbound() <-chan *datalink.Message {
	return s.l.inbound
}

func (s *Server) Done() <-chan struct{} {
	return s.rpcServer.Done()
}

type listener struct {
	outbound chan *datalink.Confirmation
	inbound  chan *datalink.Message
	datalink.UnimplementedDataLinkServer
	state *nodeDFA
}

func newListener(state *nodeDFA, buffer int) *listener {
	return &listener{
		outbound: make(chan *datalink.Confirmation, buffer),
		inbound:  make(chan *datalink.Message, buffer),
		state:    state,
	}
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) Handshake(grpc.BidiStreamingServer[datalink.ClientHandshakeMsg, datalink.ServerHandshakeMsg]) error {
	return status.Error(codes.Unimplemented, "method Handshake not implemented")
}

func (l *listener) Replicate(s grpc.BidiStreamingServer[datalink.Message, datalink.Confirmation]) error {
	l.state.emit(predecessorConnect)
	defer l.state.emit(predecessorDisconnect)
	if p, ok := peer.FromContext(s.Context()); ok {
		log.Println("New node connected:", p.Addr.String())
	} else {
		log.Println("New node connected")
	}
	ctx := s.Context()
	supervisor := stream.NewSupervisor(l.outbound, l.inbound)
	defer func() {
		if supervisor.DroppedMessage() != nil {
			panic("dropped message")
		}
	}()
	return supervisor.Run(ctx, s)
}
