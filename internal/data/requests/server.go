package requests

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/storage"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
)

type listener struct {
	db *storage.AppDatabase
	razpravljalnica.UnimplementedMessageBoardServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterMessageBoardServer(grpcServer, l)
}

type Server struct {
	rpcServer *rpc.Server
}

func NewServer(ctx context.Context, addr string) *Server {
	l := &listener{}
	s := rpc.NewServer(ctx, l, addr)
	return &Server{
		rpcServer: s,
	}
}
