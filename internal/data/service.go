package data

import (
	"context"
	"seminarska/internal/data/chain"
	"seminarska/internal/data/config"
	"seminarska/internal/data/control"
	"seminarska/internal/data/requests"
	"seminarska/internal/data/storage"
)

type Service struct {
	requestsServer *requests.Server
	database       *storage.AppDatabase
	node           *chain.Node
	control        *control.Server
	ctx            context.Context
}

func NewService(ctx context.Context, config config.NodeConfig) *Service {
	database := storage.NewAppDatabase()
	node := chain.NewNode(
		ctx,
		database.ReplicationHandler(),
		database.ReplicationHandler(),
		database,
		config.ChainListenerAddress,
	)
	s := &Service{
		ctx:            ctx,
		database:       database,
		requestsServer: requests.NewServer(ctx, database, config.ServiceAddress),
		control:        control.NewServer(ctx, config.ControlListenerAddress, node),
		node:           node,
	}
	return s
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
