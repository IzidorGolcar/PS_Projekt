package rpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type GrpcService interface {
	Register(grpcServer *grpc.Server)
}

type Server struct {
	ctx     context.Context
	addr    string
	s       *grpc.Server
	service GrpcService
	done    chan struct{}
}

func NewServer(
	ctx context.Context,
	service GrpcService,
	addr string,
) *Server {
	s := &Server{
		ctx:     ctx,
		s:       grpc.NewServer(),
		service: service,
		addr:    addr,
		done:    make(chan struct{}),
	}
	s.run()
	return s
}

func (s *Server) run() {
	s.service.Register(s.s)
	go s.serve()
	go s.handleShutdown()
}

func (s *Server) handleShutdown() {
	<-s.ctx.Done()
	s.s.GracefulStop()
}

func (s *Server) serve() {
	defer close(s.done)
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := s.s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}
