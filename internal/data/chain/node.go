package chain

import (
	"context"
	"log"
	"seminarska/internal/data/config"
	"seminarska/internal/data/storage/entities"
	"time"
)

type Node struct {
	next        *Client
	tail        *Client
	chainServer *Server
}

func datalinkContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func (n *Node) Forward(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Write(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Delete(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Delete(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Update(entity entities.Entity) error {
	if n.next == nil {
		return nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	_, err := n.next.Update(ctx, entities.EntityToDatalink(entity))
	return err
}

func (n *Node) Compare(entity entities.Entity) (bool, error) {
	if n.next == nil {
		return true, nil
	}
	ctx, cancel := datalinkContext()
	defer cancel()
	cmp, err := n.next.Compare(ctx, entities.EntityToDatalink(entity))
	if err != nil {
		return false, err
	}
	return cmp.Equal, nil
}

func NewNode(ctx context.Context, config config.NodeConfig) *Node {
	return &Node{
		chainServer: NewServer(ctx, config.ChainListenerAddress),
		next:        NewClient(ctx, config.ChainTargetAddress),
		tail:        NewClient(ctx, config.TailAddress),
	}
}

func (n *Node) Await(ctx context.Context) {
loop:
	for i := 3; i > 0; i-- {
		select {
		case <-ctx.Done():
			log.Println("failed to gracefully shutdown")
			break loop
		case <-n.chainServer.Done():
			continue
		case <-n.next.Done():
			continue
		case <-n.tail.Done():
			continue

		}
	}
	<-n.chainServer.Done()
}
