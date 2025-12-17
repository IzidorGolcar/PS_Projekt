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
	node           *chain.Node
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewService(ctx context.Context, config config.NodeConfig) *Service {
	serverCtx, cancel := context.WithCancel(ctx)
	database := storage.NewAppDatabase()
	n := &Service{
		ctx:            serverCtx,
		cancel:         cancel,
		database:       database,
		node:           chain.NewNode(serverCtx, database.Chain(), config.ChainListenerAddress),
		requestsServer: requests.NewServer(serverCtx, config.ServiceAddress),
	}
	return n
}

func (n *Service) Close() {
	n.cancel()
}

func (n *Service) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-n.node.Done()
		<-n.requestsServer.Done()
	}()
	return done
}
