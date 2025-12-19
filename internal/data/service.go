package data

import (
	"context"
	"log"
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
	done           chan struct{}
}

func NewService(ctx context.Context, config config.NodeConfig) *Service {
	database := storage.NewAppDatabase()
	s := &Service{
		ctx:            ctx,
		database:       database,
		requestsServer: requests.NewServer(ctx, database, config.ServiceAddress),
		control:        control.NewServer(ctx, config.ControlListenerAddress),
		node: chain.NewNode(
			ctx,
			database.ReplicationHandler(),
			database.ReplicationHandler(),
			database,
			config.ChainListenerAddress,
		),
		done: make(chan struct{}),
	}
	go s.run()
	return s
}

func (n *Service) run() {
	defer close(n.done)
	for {
		select {
		case <-n.ctx.Done():
			return
		case cmd := <-n.control.Commands():
			n.execute(cmd)
		}
	}
}

func (n *Service) execute(cmd control.Command) {
	switch cmd := cmd.(type) {
	case control.SwitchSuccessorCommand:
		err := n.node.SetNextNode(cmd.NewAddress)
		if err != nil {
			log.Println("Failed to switch successor:", err)
		}
	}
}

func (n *Service) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-n.done
		<-n.node.Done()
		<-n.requestsServer.Done()
	}()
	return done
}
