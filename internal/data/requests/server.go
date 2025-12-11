package requests

import (
	"context"
	"seminarska/internal/common/rpc"
)

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
