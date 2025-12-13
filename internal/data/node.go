package data

import (
	"context"
	"log"
	"seminarska/internal/data/chain"
	"seminarska/internal/data/config"
	"seminarska/internal/data/requests"
	"seminarska/internal/data/storage"
	"seminarska/proto/datalink"
)

type Node struct {
	requestsServer *requests.Server
	database       *storage.Database
	chain          *chainedNode

	ctx    context.Context
	cancel context.CancelFunc
}

type chainedNode struct {
	chainClient *chain.Client
	tailClient  *chain.Client
	chainServer *chain.Server
}

func newChainedNode(ctx context.Context, config config.NodeConfig) *chainedNode {
	return &chainedNode{
		chainServer: chain.NewServer(ctx, config.ChainListenerAddress),
		chainClient: chain.NewClient(ctx, config.ChainTargetAddress),
		tailClient:  chain.NewClient(ctx, config.TailAddress),
	}
}

func (n *chainedNode) ForwardRequest(record *datalink.Record) error {
	return n.chainClient.WriteData(record)
}

func (n *chainedNode) IsSynced(record *datalink.Record) (bool, error) {
	return n.tailClient.Sync(record)
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
		database:       storage.NewDatabase(&chainedNode{}),
		chain:          newChainedNode(serverCtx, config),
		requestsServer: requests.NewServer(serverCtx, config.ServiceAddress),
	}
	return n
}

func (n *Node) Await(ctx context.Context) {
	n.chain.await(ctx)
}
