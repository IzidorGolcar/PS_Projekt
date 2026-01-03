package requests

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/internal/data/storage"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
)

type listener struct {
	subToken string
	db       *storage.AppDatabase
	razpravljalnica.UnimplementedMessageBoardServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterMessageBoardServer(grpcServer, l)
}

type Server struct {
	rpcServer *rpc.Server
}

func NewServer(
	ctx context.Context,
	database *storage.AppDatabase,
	addr string,
	subToken string,
) *Server {
	l := &listener{db: database, subToken: subToken}
	s := rpc.NewServer(ctx, l, addr)
	return &Server{
		rpcServer: s,
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.rpcServer.Done()
}
