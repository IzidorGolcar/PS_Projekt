package control

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CommandHandler interface {
	SetNextNode(address string) error
	SetRole(role controllink.NodeRole) error
}

type Server struct {
	rpcServer *rpc.Server
}

func NewServer(ctx context.Context, addr string, h CommandHandler) *Server {
	l := &listener{handler: h}
	return &Server{rpcServer: rpc.NewServer(ctx, l, addr)}
}

type listener struct {
	controllink.UnimplementedControlServiceServer
	handler CommandHandler
}

func (l *listener) Register(grpcServer *grpc.Server) {
	controllink.RegisterControlServiceServer(grpcServer, l)
}

func (l *listener) SwitchSuccessor(
	_ context.Context,
	req *controllink.SwitchSuccessorCommand,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, l.handler.SetNextNode(req.GetAddress())
}

func (l *listener) SwitchRole(
	_ context.Context,
	req *controllink.SwitchRoleCommand,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, l.handler.SetRole(req.GetRole())
}

func (l *listener) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
