package chain

import (
	"context"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
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

func (l *listener) Replicate(stream grpc.BidiStreamingServer[datalink.Message, datalink.Confirmation]) error {
	l.state.emit(predecessorConnect)
	defer l.state.emit(predecessorDisconnect)
	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Println("New node connected:", p.Addr.String())
	} else {
		log.Println("New node connected")
	}
	ctx := stream.Context()
	supervisor := NewStreamSupervisor(l.outbound, l.inbound)
	defer func() {
		if supervisor.DroppedMessage() != nil {
			panic("dropped message")
		}
	}()
	return supervisor.Run(ctx, stream)
}
