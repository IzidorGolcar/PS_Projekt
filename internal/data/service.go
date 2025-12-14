package data

import (
	"context"
	"seminarska/internal/data/chain"
	"seminarska/internal/data/config"
	"seminarska/internal/data/requests"
	"seminarska/internal/data/storage"
)

type Service struct {
	requestsServer *requests.Server
	database       *storage.AppDatabase
	chain          *chain.Node
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewService(ctx context.Context, config config.NodeConfig) *Service {
	serverCtx, cancel := context.WithCancel(ctx)
	chainNode := chain.NewNode(serverCtx, config)
	n := &Service{
		ctx:            serverCtx,
		cancel:         cancel,
		database:       storage.NewAppDatabase(chainNode),
		chain:          chainNode,
		requestsServer: requests.NewServer(serverCtx, config.ServiceAddress),
	}
	return n
}

func (n *Service) Await(ctx context.Context) {
	n.chain.Await(ctx)
}
