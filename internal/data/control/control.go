package control

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Command interface{}

type SwitchSuccessorCommand struct {
	NewAddress string
}

type Server struct {
	rpcServer *rpc.Server
	commands  chan Command
}

func (s *Server) Commands() <-chan Command {
	return s.commands
}

func NewServer(ctx context.Context, addr string) *Server {
	commands := make(chan Command, 100)
	l := &listener{commands: commands}
	return &Server{
		rpcServer: rpc.NewServer(ctx, l, addr),
		commands:  commands,
	}
}

type listener struct {
	controllink.UnimplementedControlServiceServer
	commands chan<- Command
}

func (l *listener) Register(grpcServer *grpc.Server) {
	controllink.RegisterControlServiceServer(grpcServer, l)
}

func (l *listener) SwitchSuccessor(
	ctx context.Context,
	req *controllink.SwitchSuccessorCommand,
) (*emptypb.Empty, error) {
	select {
	case l.commands <- SwitchSuccessorCommand{NewAddress: req.GetAddress()}:
		return &emptypb.Empty{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
