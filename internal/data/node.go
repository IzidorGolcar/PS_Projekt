package data

import (
	"context"
	"log"
	"seminarska/internal/data/chain"
	"seminarska/internal/data/config"
	"seminarska/internal/data/requests"
	"seminarska/internal/data/storage"
	"seminarska/internal/data/storage/entities"
	"time"
)

type Node struct {
	requestsServer *requests.Server
	database       *storage.AppDatabase
	chain          *chainedNode

	ctx    context.Context
	cancel context.CancelFunc
}

type chainedNode struct {
	chainClient *chain.Client
	tailClient  *chain.Client
	chainServer *chain.Server
}

func datalinkContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func (n *chainedNode) Forward(entity entities.Entity) error {
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.chainClient.Write(ctx, entity.ToDatalinkRecord())
	return err
}

func (n *chainedNode) Delete(entity entities.Entity) error {
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.chainClient.Delete(ctx, entity.ToDatalinkRecord())
	return err
}

func (n *chainedNode) Update(entity entities.Entity) error {
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.chainClient.Update(ctx, entity.ToDatalinkRecord())
	return err
}

func (n *chainedNode) Compare(entity entities.Entity) (bool, error) {
	ctx, cancel := datalinkContext()
	defer cancel()
	cmp, err := n.chainClient.Compare(ctx, entity.ToDatalinkRecord())
	if err != nil {
		return false, err
	}
	return cmp.Equal, nil
}

func newChainedNode(ctx context.Context, config config.NodeConfig) *chainedNode {
	return &chainedNode{
		chainServer: chain.NewServer(ctx, config.ChainListenerAddress),
		chainClient: chain.NewClient(ctx, config.ChainTargetAddress),
		tailClient:  chain.NewClient(ctx, config.TailAddress),
	}
}

func (n *chainedNode) await(ctx context.Context) {
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			log.Println("failed to gracefully shutdown")
		case <-n.chainServer.Done():
			continue
		case <-n.chainClient.Done():
			continue
		case <-n.tailClient.Done():
			continue

		}
	}
	<-n.chainServer.Done()
}

func NewNode(ctx context.Context, config config.NodeConfig) *Node {
	serverCtx, cancel := context.WithCancel(ctx)
	n := &Node{
		ctx:            serverCtx,
		cancel:         cancel,
		database:       storage.NewAppDatabase(&chainedNode{}),
		chain:          newChainedNode(serverCtx, config),
		requestsServer: requests.NewServer(serverCtx, config.ServiceAddress),
	}
	return n
}

func (n *Node) Await(ctx context.Context) {
	n.chain.await(ctx)
}
