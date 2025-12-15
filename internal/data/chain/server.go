package chain

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/datalink"

	"google.golang.org/grpc"
)

type Server struct {
	l         *listener
	rpcServer *rpc.Server
}

func NewServer(ctx context.Context, addr string, buffer int) *Server {
	l := newListener(buffer)
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
}

func newListener(buffer int) *listener {
	return &listener{
		outbound: make(chan *datalink.Confirmation, buffer),
		inbound:  make(chan *datalink.Message, buffer),
	}
}

func (l *listener) Register(grpcServer *grpc.Server) {
	datalink.RegisterDataLinkServer(grpcServer, l)
}

func (l *listener) Replicate(stream grpc.BidiStreamingServer[datalink.Message, datalink.Confirmation]) error {
	ctx := stream.Context()
	supervisor := NewStreamSupervisor(l.outbound, l.inbound)
	defer func() {
		if supervisor.DroppedMessage() != nil {
			panic("dropped message")
		}
	}()
	return supervisor.Run(ctx, stream)
}
