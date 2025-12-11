package requests

import (
	"seminarska/internal/data/storage"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
)

type listener struct {
	db *storage.Database
	razpravljalnica.UnimplementedMessageBoardServer
}

func (l *listener) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterMessageBoardServer(grpcServer, l)
}
